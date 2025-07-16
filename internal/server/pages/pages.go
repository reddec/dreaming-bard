package pages

import (
	"archive/zip"
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-shiori/go-epub"

	"github.com/reddec/dreaming-bard/internal/dbo"
	"github.com/reddec/dreaming-bard/internal/dreamwriter"
	"github.com/reddec/dreaming-bard/internal/server/views"
	"github.com/reddec/dreaming-bard/internal/utils/lifecycle"
	"github.com/reddec/dreaming-bard/internal/utils/mark"
)

type pageMeta struct {
	ID      int64     `yaml:"id,omitempty"`
	Created time.Time `yaml:"created,omitempty"`
	Updated time.Time `yaml:"updated,omitempty"`
	Num     int64     `yaml:"num,omitempty"`
	Summary string    `yaml:"summary,omitempty"`
}

func New(dreamer *dreamwriter.DreamWriter) *Pages {
	mux := http.NewServeMux()
	srv := &Pages{
		Handler:  mux,
		dreamer:  dreamer,
		showHelp: dbo.NewPref[bool](dreamer.DB(), "help_page", true),
	}
	mux.HandleFunc("GET /", srv.list)
	mux.HandleFunc("GET /new", srv.wizard)
	mux.HandleFunc("POST /new", srv.wizard) // prefill
	mux.HandleFunc("POST /{$}", srv.create)
	mux.HandleFunc("POST /help", views.BoolHandler(srv.showHelp))
	mux.HandleFunc("GET /pages.zip", srv.exportAll)
	mux.HandleFunc("GET /import", srv.importWizard)
	mux.HandleFunc("POST /import", srv.importAll)
	mux.HandleFunc("GET /epub", srv.epubWizard)
	mux.HandleFunc("POST /epub", srv.epub)
	mux.HandleFunc("GET /{pageID}/", srv.index)
	mux.HandleFunc("POST /{pageID}/{$}", srv.update)
	mux.HandleFunc("POST /{pageID}/move", srv.move)
	mux.HandleFunc("DELETE /{pageID}/", srv.delete)
	mux.HandleFunc("POST /{pageID}/generate-summary", srv.generateSummary)

	return srv
}

type Pages struct {
	http.Handler
	dreamer  *dreamwriter.DreamWriter
	showHelp *dbo.Pref[bool]
}

func (srv *Pages) Run(ctx context.Context) error {
	// TODO: background tasks
	return nil
}

func (srv *Pages) list(w http.ResponseWriter, r *http.Request) {
	pages, err := srv.dreamer.DB().ListPages(r.Context())
	if err != nil {
		views.RenderError(w, err)
		return
	}
	help, err := srv.showHelp.Get(r.Context())
	if err != nil {
		views.RenderError(w, err)
		return
	}
	viewList().HTML(w, listParams{
		Pages:    pages,
		ShowHelp: help,
	})
}

func (srv *Pages) exportAll(w http.ResponseWriter, r *http.Request) {
	pages, err := srv.dreamer.DB().ListPages(r.Context())
	if err != nil {
		views.RenderError(w, err)
		return
	}
	w.Header().Set("Content-Type", "application/zip")
	w.Header().Set("Content-Disposition", "attachment; filename=pages.zip")
	out := zip.NewWriter(w)

	for _, page := range pages {
		md := mark.Document[pageMeta]{
			Metadata: pageMeta{
				ID:      page.ID,
				Created: page.CreatedAt,
				Updated: page.UpdatedAt,
				Num:     page.Num,
				Summary: page.Summary,
			},
			Content: page.Content,
		}
		doc, err := md.Build()
		if err != nil {
			views.RenderError(w, err)
			return
		}

		f, err := out.Create(fmt.Sprintf("%04d.md", page.Num))
		if err != nil {
			views.RenderError(w, err)
			return
		}
		_, err = f.Write([]byte(doc))
		if err != nil {
			views.RenderError(w, err)
			return
		}
	}
	err = out.Close()
	if err != nil {
		views.RenderError(w, err)
		return
	}
}

func (srv *Pages) importWizard(w http.ResponseWriter, r *http.Request) {
	viewImport().HTML(w, importParams{})
}

func (srv *Pages) importAll(w http.ResponseWriter, r *http.Request) {
	err := lifecycle.Import[pageMeta](r, func(doc mark.Document[pageMeta], name string) error {
		if _, err := srv.dreamer.DB().CreatePage(r.Context(), dbo.CreatePageParams{
			Summary: doc.Metadata.Summary,
			Content: doc.Content,
		}); err != nil {
			return fmt.Errorf("import page %q: %w", name, err)
		}
		if doc.Metadata.Num > 0 {
			if err := srv.dreamer.DB().UpdatePageNum(r.Context(), doc.Metadata.ID, doc.Metadata.Num); err != nil {
				return fmt.Errorf("move page %q: %w", name, err)
			}
		}
		return nil
	})

	if err != nil {
		views.RenderError(w, err)
		return
	}

	w.Header().Set("Location", ".")
	w.WriteHeader(http.StatusSeeOther)
}

func (srv *Pages) move(w http.ResponseWriter, r *http.Request) {
	pageID, _ := strconv.ParseInt(r.PathValue("pageID"), 10, 64)
	pageNum, _ := strconv.ParseInt(r.FormValue("num"), 10, 64)
	if err := srv.dreamer.DB().UpdatePageNum(r.Context(), pageID, pageNum); err != nil {
		views.RenderError(w, err)
		return
	}
	w.Header().Set("Location", ".")
	w.WriteHeader(http.StatusSeeOther)
}

func (srv *Pages) index(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("pageID"), 10, 64)
	if err != nil {
		views.RenderError(w, err)
		return
	}
	page, err := srv.dreamer.DB().GetPage(r.Context(), id)
	if err != nil {
		views.RenderError(w, err)
		return
	}

	if isEdit, _ := strconv.ParseBool(r.FormValue("edit")); isEdit {
		viewEdit().HTML(w, editParams{
			Page: page,
		})
	} else {
		roles, err := srv.dreamer.DB().ListRoles(r.Context())
		if err != nil {
			views.RenderError(w, err)
			return
		}
		viewIndex().HTML(w, indexParams{
			Roles: roles,
			Page:  page,
		})
	}
}

func (srv *Pages) wizard(w http.ResponseWriter, r *http.Request) {
	viewNew().HTML(w, newParams{
		Prefill: r.FormValue("content"),
	})
}

func (srv *Pages) create(w http.ResponseWriter, r *http.Request) {
	summary := strings.TrimSpace(r.FormValue("summary"))
	content := strings.TrimSpace(r.FormValue("content"))
	page, err := srv.dreamer.DB().CreatePage(r.Context(), dbo.CreatePageParams{
		Summary: summary,
		Content: content,
	})
	if err != nil {
		views.RenderError(w, err)
		return
	}
	w.Header().Set("Location", strconv.FormatInt(page.ID, 10)+"/?"+r.URL.RawQuery)
	w.WriteHeader(http.StatusSeeOther)
}

func (srv *Pages) update(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("pageID"), 10, 64)
	if err != nil {
		views.RenderError(w, err)
		return
	}
	summary := strings.TrimSpace(r.FormValue("summary"))
	content := strings.TrimSpace(r.FormValue("content"))
	err = srv.dreamer.DB().UpdatePage(r.Context(), dbo.UpdatePageParams{
		Summary: summary,
		Content: content,
		ID:      id,
	})
	if err != nil {
		views.RenderError(w, err)
		return
	}
	w.Header().Set("Location", "?"+r.URL.RawQuery)
	w.WriteHeader(http.StatusSeeOther)
}

func (srv *Pages) generateSummary(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("pageID"), 10, 64)
	if err != nil {
		views.RenderError(w, err)
		return
	}
	page, err := srv.dreamer.DB().GetPage(r.Context(), id)
	if err != nil {
		views.RenderError(w, err)
		return
	}
	roleID, _ := strconv.ParseInt(r.FormValue("role"), 10, 64)

	summary, err := srv.dreamer.Summarise(r.Context(), roleID, page.Content)
	if err != nil {
		views.RenderError(w, err)
		return
	}

	err = srv.dreamer.DB().UpdatePageSummary(r.Context(), dbo.UpdatePageSummaryParams{
		Summary: summary,
		ID:      id,
	})
	if err != nil {
		views.RenderError(w, err)
		return
	}
	w.Header().Set("Location", ".")
	w.WriteHeader(http.StatusSeeOther)
}

func (srv *Pages) delete(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("pageID"), 10, 64)
	if err != nil {
		views.RenderError(w, err)
		return
	}
	if err := srv.dreamer.DB().DeletePage(r.Context(), id); err != nil {
		views.RenderError(w, err)
		return
	}

	if views.IsHTMX(r) {
		if r.FormValue("stay") != "true" {
			w.Header().Set("HX-Redirect", "../")
		}
		w.WriteHeader(http.StatusOK)
		return
	}
	w.Header().Set("Location", "../")
	w.WriteHeader(http.StatusSeeOther)
}

func (srv *Pages) epubWizard(w http.ResponseWriter, r *http.Request) {
	viewEpub().HTML(w, epubParams{})
}

func (srv *Pages) epub(w http.ResponseWriter, r *http.Request) {
	author := strings.TrimSpace(r.FormValue("author"))
	title := strings.TrimSpace(r.FormValue("title"))

	pages, err := srv.dreamer.DB().ListPages(r.Context())
	if err != nil {
		views.RenderError(w, err)
		return
	}

	book, err := epub.NewEpub(title)
	if err != nil {
		views.RenderError(w, err)
		return
	}
	book.SetAuthor(author)

	for _, page := range pages {
		content, err := views.ConvertMarkdown(page.Content)
		if err != nil {
			views.RenderError(w, fmt.Errorf("render page %d: %w", page.Num, err))
			return
		}
		content = strings.ReplaceAll(content, "<hr>", "<hr/>") // workaround for self closed tags

		_, err = book.AddSection(content, "#"+strconv.FormatInt(page.Num, 10), "", "")
		if err != nil {
			views.RenderError(w, err)
			return
		}
	}

	w.Header().Set("Content-Type", "application/epub+zip")
	w.Header().Set("Content-Disposition", "attachment; filename="+strconv.Quote(title+".epub"))

	if _, err := book.WriteTo(w); err != nil {
		slog.Error("failed to write epub", "author", author, "title", title, "error", err)
	}
}
