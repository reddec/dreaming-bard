package contexts

import (
	"archive/zip"
	"context"
	"fmt"
	"io"
	"maps"
	"net/http"
	"path"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"

	"github.com/reddec/dreaming-bard/internal/dbo"
	"github.com/reddec/dreaming-bard/internal/dreamwriter"
	"github.com/reddec/dreaming-bard/internal/server/views"
	"github.com/reddec/dreaming-bard/internal/utils/mark"
)

func New(dreamer *dreamwriter.DreamWriter) *Contexts {
	mux := http.NewServeMux()
	srv := &Contexts{
		Handler:  mux,
		dreamer:  dreamer,
		showHelp: dbo.NewPref[bool](dreamer.DB(), "help_context", true),
	}
	mux.HandleFunc("GET /{$}", srv.list)
	mux.HandleFunc("GET /new", srv.wizard)
	mux.HandleFunc("POST /new", srv.wizard) // prefill
	mux.HandleFunc("POST /help", views.BoolHandler(srv.showHelp))
	mux.HandleFunc("GET /context.zip", srv.export)
	mux.HandleFunc("GET /import", srv.importWizard)
	mux.HandleFunc("POST /import", srv.importAll)
	mux.HandleFunc("POST /upload", srv.importIndividual)

	mux.HandleFunc("POST /{$}", srv.createContext)
	mux.HandleFunc("GET /{factID}/", srv.getContext)

	mux.HandleFunc("DELETE /{factID}/", srv.deleteContext)
	mux.HandleFunc("POST /{factID}/", srv.updateContext)
	mux.HandleFunc("POST /{factID}/archived", srv.updateArchiveState)

	return srv
}

type Contexts struct {
	http.Handler
	dreamer  *dreamwriter.DreamWriter
	showHelp *dbo.Pref[bool]
}

func (srv *Contexts) Run(ctx context.Context) error {
	// TODO: background tasks
	return nil
}

func (srv *Contexts) list(w http.ResponseWriter, r *http.Request) {
	category := r.FormValue("category")
	var facts []dbo.Context
	var err error
	if category != "" {
		facts, err = srv.dreamer.DB().ListContextsByCategory(r.Context(), category)
	} else {
		facts, err = srv.dreamer.DB().ListContexts(r.Context())
	}
	if err != nil {
		views.RenderError(w, err)
		return
	}
	categories, err := srv.dreamer.DB().ListContextCategories(r.Context())
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
		Facts:      facts,
		Categories: categories,
		Category:   category,
		ShowHelp:   help,
	})
}

func (srv *Contexts) getContext(w http.ResponseWriter, r *http.Request) {
	factID, _ := strconv.ParseInt(r.PathValue("factID"), 10, 64)
	isEdit, _ := strconv.ParseBool(r.FormValue("edit"))
	fact, err := srv.dreamer.DB().GetContext(r.Context(), factID)
	if err != nil {
		views.RenderError(w, err)
		return
	}
	if isEdit {
		viewEdit().HTML(w, editParams{
			Fact: fact,
		})
	} else {
		viewIndex().HTML(w, indexParams{
			Fact: fact,
		})
	}

}

func (srv *Contexts) wizard(w http.ResponseWriter, r *http.Request) {
	prefillContent := r.FormValue("content")

	facts, err := srv.dreamer.DB().ListContexts(r.Context())
	if err != nil {
		views.RenderError(w, err)
		return
	}

	var categories = make(map[string]struct{})
	for _, fact := range facts {
		categories[fact.Category] = struct{}{}
	}

	manualOnly, _ := strconv.ParseBool(r.FormValue("manualOnly"))
	viewNew().HTML(w, newParams{
		Facts:      facts,
		Categories: slices.Collect(maps.Keys(categories)),
		Prefill:    map[string]string{"content": prefillContent},
		ManualOnly: manualOnly,
	})
}

func (srv *Contexts) export(w http.ResponseWriter, r *http.Request) {
	facts, err := srv.dreamer.DB().ListContexts(r.Context())
	if err != nil {
		views.RenderError(w, err)
		return
	}
	w.Header().Set("Content-Type", "application/zip")
	w.Header().Set("Content-Disposition", "attachment; filename=context.zip")
	out := zip.NewWriter(w)

	for _, fact := range facts {
		md := mark.Document[factMeta]{
			Metadata: factMeta{
				ID:       fact.ID,
				Category: fact.Category,
				Created:  fact.CreatedAt,
				Updated:  fact.UpdatedAt,
				Title:    fact.Title,
			},
			Content: fact.Content,
		}

		doc, err := md.Build()
		if err != nil {
			views.RenderError(w, err)
			return
		}

		f, err := out.Create(fact.Title + ".md")
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

func (srv *Contexts) deleteContext(w http.ResponseWriter, r *http.Request) {
	factID, _ := strconv.ParseInt(r.PathValue("factID"), 10, 64)
	if err := srv.dreamer.DB().DeleteContext(r.Context(), factID); err != nil {
		views.RenderError(w, err)
		return
	}
	if views.IsHTMX(r) {
		w.Header().Set("HX-Redirect", "../")
	} else {
		w.Header().Set("Location", "../")
	}
	w.WriteHeader(http.StatusSeeOther)
}

func (srv *Contexts) importIndividual(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		views.RenderError(w, err)
		return
	}
	isRemoveTags, _ := strconv.ParseBool(r.FormValue("removeTags"))
	isRemoveInlineWiki, _ := strconv.ParseBool(r.FormValue("removeInlineWiki"))
	isRemoveWikiLinks, _ := strconv.ParseBool(r.FormValue("removeWikiLinks"))

	files := r.MultipartForm.File["file"]
	if len(files) == 0 {
		views.RenderError(w, fmt.Errorf("no files"))
		return
	}
	for _, file := range files {
		if file.Size > 10*1024*1024 {
			views.RenderError(w, fmt.Errorf("file too large"))
			return
		}
		stream, err := file.Open()
		if err != nil {
			views.RenderError(w, err)
			return
		}
		content, err := io.ReadAll(stream)
		_ = stream.Close()
		if err != nil {
			views.RenderError(w, err)
			return
		}

		doc := mark.Parse[factMeta](string(content))
		name := doc.Metadata.Title // allow overriding document title instead of a filename
		if name == "" {
			ext := filepath.Ext(file.Filename)
			name = strings.TrimSuffix(file.Filename, ext)
		}

		if isRemoveTags {
			doc.Content = removeTags(doc.Content)
		}
		if isRemoveInlineWiki {
			doc.Content = removeInlineWiki(doc.Content)
		}
		if isRemoveWikiLinks {
			doc.Content = removeWikiLinks(doc.Content)
		}

		if _, err := srv.dreamer.DB().CreateContext(r.Context(), dbo.CreateContextParams{
			Title:    name,
			Category: doc.Metadata.Category,
			Content:  doc.Content,
		}); err != nil {
			views.RenderError(w, fmt.Errorf("save doc %q: %w", name, err))
			return
		}
	}
	w.Header().Set("Location", ".")
	w.WriteHeader(http.StatusSeeOther)
}

func (srv *Contexts) importAll(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		views.RenderError(w, err)
		return
	}

	files := r.MultipartForm.File["archive"]
	if len(files) == 0 {
		views.RenderError(w, fmt.Errorf("no archive"))
		return
	}
	arch := files[0]
	f, err := arch.Open()
	if err != nil {
		views.RenderError(w, err)
		return
	}
	defer f.Close()

	reader, err := zip.NewReader(f, arch.Size)
	if err != nil {
		views.RenderError(w, err)
		return
	}
	for _, file := range reader.File {
		stream, err := file.Open()
		if err != nil {
			views.RenderError(w, err)
			return
		}

		content, err := io.ReadAll(stream)
		_ = stream.Close()
		if err != nil {
			views.RenderError(w, err)
			return
		}

		doc := mark.Parse[factMeta](string(content))
		name := doc.Metadata.Title // allow overriding document title instead of a filename
		if name == "" {
			base := path.Base(file.Name)
			ext := filepath.Ext(base)
			name = strings.TrimSuffix(base, ext)
		}

		if _, err := srv.dreamer.DB().CreateContext(r.Context(), dbo.CreateContextParams{
			Title:    name,
			Category: doc.Metadata.Category,
			Content:  doc.Content,
		}); err != nil {
			views.RenderError(w, fmt.Errorf("import doc %q: %w", file.Name, err))
			return
		}
	}
	w.Header().Set("Location", ".")
	w.WriteHeader(http.StatusSeeOther)
}

func (srv *Contexts) importWizard(w http.ResponseWriter, r *http.Request) {
	viewImport().HTML(w, importParams{})
}

func (srv *Contexts) createContext(w http.ResponseWriter, r *http.Request) {
	name := strings.TrimSpace(r.FormValue("name"))

	category := strings.TrimSpace(r.FormValue("category"))
	content := strings.TrimSpace(r.FormValue("content"))
	doc, err := srv.dreamer.DB().CreateContext(r.Context(), dbo.CreateContextParams{
		Title:    name,
		Category: category,
		Content:  content,
	})
	if err != nil {
		views.RenderError(w, err)
		return
	}
	w.Header().Set("Location", strconv.FormatInt(doc.ID, 10)+"/")
	w.WriteHeader(http.StatusSeeOther)
}

func (srv *Contexts) updateContext(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(r.PathValue("factID"), 10, 64)
	name := strings.TrimSpace(r.FormValue("name"))
	category := strings.TrimSpace(r.FormValue("category"))
	content := strings.TrimSpace(r.FormValue("content"))
	err := srv.dreamer.DB().UpdateContext(r.Context(), dbo.UpdateContextParams{
		Title:    name,
		Category: category,
		Content:  content,
		ID:       id,
	})
	if err != nil {
		views.RenderError(w, err)
		return
	}
	w.Header().Set("Location", ".")
	w.WriteHeader(http.StatusSeeOther)
}

func (srv *Contexts) updateArchiveState(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(r.PathValue("factID"), 10, 64)
	archived, _ := strconv.ParseBool(r.FormValue("archived"))
	err := srv.dreamer.DB().UpdateContextArchivedStatus(r.Context(), dbo.UpdateContextArchivedStatusParams{
		Archived: archived,
		ID:       id,
	})
	if err != nil {
		views.RenderError(w, err)
		return
	}
	if views.IsHTMX(r) {
		//w.Header().Set("HX-Redirect", ".")
		w.Header().Set("Location", r.Header.Get("hx-current-url"))
		w.WriteHeader(http.StatusSeeOther)
	} else {
		w.Header().Set("Location", ".")
		w.WriteHeader(http.StatusSeeOther)
	}
}

func removeTags(content string) string {
	content = regexp.MustCompile(`#[a-zA-Z0-9_-]+`).ReplaceAllString(content, "")
	content = regexp.MustCompile(` +`).ReplaceAllString(content, " ")
	return strings.TrimSpace(content)
}

func removeInlineWiki(content string) string {
	return regexp.MustCompile(`!\[\[([^]]+)]]`).ReplaceAllString(content, "See $1")
}

func removeWikiLinks(content string) string {
	return regexp.MustCompile(`\[\[([^]]+)]]`).ReplaceAllString(content, "$1")
}
