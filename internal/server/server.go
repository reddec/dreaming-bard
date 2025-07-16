package server

import (
	"context"
	"embed"
	"net/http"
	"strings"

	"github.com/sourcegraph/conc/pool"

	"github.com/reddec/dreaming-bard/internal/dbo"
	"github.com/reddec/dreaming-bard/internal/dreamwriter"
	"github.com/reddec/dreaming-bard/internal/llm"
	"github.com/reddec/dreaming-bard/internal/server/blueprint"
	"github.com/reddec/dreaming-bard/internal/server/chats"
	"github.com/reddec/dreaming-bard/internal/server/contexts"
	"github.com/reddec/dreaming-bard/internal/server/home"
	"github.com/reddec/dreaming-bard/internal/server/pages"
	"github.com/reddec/dreaming-bard/internal/server/prompts"
	"github.com/reddec/dreaming-bard/internal/server/roles"
)

//go:generate ./gen.py

//go:embed all:static
var static embed.FS

func New(db *dbo.Queries, providerConfig llm.Provider) *Server {
	srv := &Server{
		providerConfig: providerConfig,
		dreamer:        dreamwriter.NewDreamWriter(db, providerConfig),
	}

	mux := http.NewServeMux()

	// pages
	srv.pages = pages.New(srv.dreamer)
	subRouter(mux, "/pages", noCache(srv.pages))

	// chats
	srv.chats = chats.New(srv.dreamer)
	subRouter(mux, "/chats", noCache(srv.chats))

	// home
	srv.home = home.New(srv.dreamer)
	subRouter(mux, "/", noCache(srv.home))

	// contexts
	srv.contexts = contexts.New(srv.dreamer)
	subRouter(mux, "/context", noCache(srv.contexts))

	// prompts
	srv.prompts = prompts.New(srv.dreamer)
	subRouter(mux, "/prompts", noCache(srv.prompts))

	// roles
	srv.roles = roles.New(srv.dreamer)
	subRouter(mux, "/roles", noCache(srv.roles))

	// blueprint
	srv.blueprint = blueprint.New(srv.dreamer, srv.chats)
	subRouter(mux, "/blueprints", noCache(srv.blueprint))

	mux.Handle("GET /static/", http.FileServerFS(static))
	srv.Handler = mux
	return srv
}

type Server struct {
	pages          *pages.Pages
	chats          *chats.Chats
	blueprint      *blueprint.Blueprint
	roles          *roles.Roles
	contexts       *contexts.Contexts
	prompts        *prompts.Prompts
	home           *home.Home
	providerConfig llm.Provider
	dreamer        *dreamwriter.DreamWriter
	http.Handler
}

func (s *Server) Run(ctx context.Context) error {
	wg := pool.New().WithContext(ctx).WithCancelOnError()
	wg.Go(s.chats.Run)
	wg.Go(s.roles.Run)
	wg.Go(s.blueprint.Run)
	wg.Go(s.prompts.Run)
	wg.Go(s.pages.Run)
	wg.Go(s.home.Run)
	wg.Go(s.contexts.Run)
	return wg.Wait()
}

func subRouter(mux *http.ServeMux, prefix string, handler http.Handler) {
	if prefix == "" {
		prefix = "/"
	}
	if prefix == "/" {
		mux.Handle("/", handler)
		return
	}
	prefix = "/" + strings.Trim(prefix, "/")
	mux.Handle(prefix+"/", http.StripPrefix(prefix, handler))
}

func noCache(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
		w.Header().Set("Pragma", "no-cache")
		w.Header().Set("Expires", "0")
		h.ServeHTTP(w, r)
	})
}
