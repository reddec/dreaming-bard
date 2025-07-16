package commands

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/NYTimes/gziphandler"
	"github.com/alexedwards/scs/v2"
	oidclogin "github.com/reddec/oidc-login"
	"github.com/rs/cors"
	"github.com/sourcegraph/conc/pool"

	"github.com/reddec/dreaming-bard/internal/dbo"
	"github.com/reddec/dreaming-bard/internal/llm"
	"github.com/reddec/dreaming-bard/internal/server"
	"github.com/reddec/dreaming-bard/internal/utils/session"
)

type ServerCommand struct {
	CORS            bool         `help:"Enable CORS" env:"CORS"`
	Bind            string       `help:"Binding address" env:"BIND" default:":8080"`
	DisableGZIP     bool         `help:"Disable gzip compression for HTTP" env:"DISABLE_GZIP"`
	ParallelWorkers int          `help:"Number of parallel workers (chats)" env:"PARALLEL_WORKERS" default:"1"`
	Provider        llm.Provider `embed:"" prefix:"provider-" envprefix:"PROVIDER_"`
	OIDC            struct {
		Enabled      bool          `help:"Enable OIDC" env:"ENABLED"`
		Issuer       string        `help:"Issuer URL" env:"ISSUER"`
		ClientID     string        `help:"Client ID" env:"CLIENT_ID"`
		ClientSecret string        `help:"Client secret" env:"CLIENT_SECRET"`
		GC           time.Duration `help:"GC interval for expired sessions" env:"GC" default:"5m"`
	} `embed:"" prefix:"oidc-" envprefix:"OIDC_"`
	TLS struct {
		Enabled  bool   `help:"Enable TLS" env:"ENABLED"`
		KeyFile  string `help:"Key file" env:"KEY" default:"/etc/tls/tls.key"`
		CertFile string `help:"Certificate file" env:"CERT" default:"/etc/tls/tls.crt"`
	} `embed:"" prefix:"tls-" envprefix:"TLS_"`
}

func (cmd *ServerCommand) Run() error {

	db, err := dbo.NewFromFile("db.sqlite3")
	if err != nil {
		return fmt.Errorf("create db: %w", err)
	}
	defer db.Close()

	srv := server.New(db, cmd.Provider)

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill)
	defer cancel()

	wg := pool.New().WithContext(ctx).WithCancelOnError()

	for i := 0; i < cmd.ParallelWorkers; i++ {
		wg.Go(srv.Run)
	}

	var handler http.Handler = srv
	if cmd.CORS {
		handler = cors.AllowAll().Handler(handler)
	}
	if cmd.OIDC.Enabled {

		sessionStore := session.NewDBSession(db)
		wg.Go(func(ctx context.Context) error {
			ticker := time.NewTicker(cmd.OIDC.GC)
			defer ticker.Stop()
			for {
				select {
				case <-ctx.Done():
					return nil
				case <-ticker.C:
					if err := sessionStore.GC(ctx); err != nil {
						slog.Error("failed to delete expired sessions", "error", err)
					}
				}
			}
		})
		manager := scs.New()
		manager.Store = sessionStore

		auth, err := oidclogin.New(ctx, oidclogin.Config{
			IssuerURL:    cmd.OIDC.Issuer,
			ClientID:     cmd.OIDC.ClientID,
			ClientSecret: cmd.OIDC.ClientSecret,
			Logger: oidclogin.LoggerFunc(func(level oidclogin.Level, msg string) {
				switch level {
				case oidclogin.LogInfo:
					slog.Info("oidc login", "message", msg)
				case oidclogin.LogWarn:
					slog.Warn("oidc login", "message", msg)
				case oidclogin.LogError:
					slog.Error("oidc login", "message", msg)
				default:
					slog.Info("oidc login", "level", level, "message", msg)
				}
			}),
			SessionManager: manager,
		})
		if err != nil {
			return fmt.Errorf("create oidc login: %w", err)
		}
		mux := http.NewServeMux()
		mux.Handle("/", auth.Secure(handler))
		mux.Handle(oidclogin.Prefix, auth)
		handler = http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			SetOWASPHeaders(writer)
			mux.ServeHTTP(writer, request)
		})
		slog.Info("OIDC enabled", "issuer", cmd.OIDC.Issuer, "client_id", cmd.OIDC.ClientID)
	}

	if !cmd.DisableGZIP {
		handler = gziphandler.GzipHandler(handler)
		slog.Info("gzip compression enabled")
	}

	httpServer := &http.Server{
		Addr:    cmd.Bind,
		Handler: handler,
	}

	wg.Go(func(ctx context.Context) error {
		<-ctx.Done()
		sub, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := httpServer.Shutdown(sub); err != nil {
			slog.Error("shutdown failed", "error", err)
		}
		return nil
	})

	wg.Go(func(_ context.Context) error {
		slog.Info("server started", "bind", cmd.Bind, "provider", cmd.Provider.Type)
		if cmd.TLS.Enabled {
			return httpServer.ListenAndServeTLS(cmd.TLS.CertFile, cmd.TLS.KeyFile)
		}

		return httpServer.ListenAndServe()

	})

	return wg.Wait()
}

func SetOWASPHeaders(writer http.ResponseWriter) {
	writer.Header().Set("X-Frame-Options", "DENY") // helps with click hijacking
	writer.Header().Set("X-XSS-Protection", "1")
	writer.Header().Set("X-Content-Type-Options", "nosniff")                  // helps with content-type substitution
	writer.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin") // disables cross-origin requests
}
