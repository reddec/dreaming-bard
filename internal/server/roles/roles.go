package roles

import (
	"archive/zip"
	"context"
	"fmt"
	"net/http"
	"path"
	"strconv"
	"strings"

	"github.com/reddec/dreaming-bard/internal/common"
	"github.com/reddec/dreaming-bard/internal/dbo"
	"github.com/reddec/dreaming-bard/internal/dreamwriter"
	"github.com/reddec/dreaming-bard/internal/server/views"
	"github.com/reddec/dreaming-bard/internal/utils/lifecycle"
	"github.com/reddec/dreaming-bard/internal/utils/mark"
)

func New(dreamer *dreamwriter.DreamWriter) *Roles {
	mux := http.NewServeMux()
	srv := &Roles{
		Handler:  mux,
		dreamer:  dreamer,
		showHelp: dbo.NewPref[bool](dreamer.DB(), "help_role", true),
	}
	mux.HandleFunc("GET /{$}", srv.list)
	mux.HandleFunc("GET /export", srv.exportAll)
	mux.HandleFunc("GET /import", srv.importWizard)
	mux.HandleFunc("POST /import", srv.importAll)
	mux.HandleFunc("POST /{$}", srv.create)
	mux.HandleFunc("GET /new", srv.wizard)
	mux.HandleFunc("POST /new", srv.wizard)
	mux.HandleFunc("POST /help", views.BoolHandler(srv.showHelp))
	mux.HandleFunc("GET /{roleID}/", srv.get)
	mux.HandleFunc("GET /{roleID}/export", srv.exportOne)
	mux.HandleFunc("POST /{roleID}/", srv.update)
	mux.HandleFunc("DELETE /{roleID}/", srv.delete)
	return srv
}

type Roles struct {
	http.Handler
	dreamer  *dreamwriter.DreamWriter
	showHelp *dbo.Pref[bool]
}

func (srv *Roles) Run(_ context.Context) error {
	// TODO: background tasks
	return nil
}

func (srv *Roles) list(w http.ResponseWriter, r *http.Request) {
	roles, err := srv.dreamer.DB().ListRoles(r.Context())
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
		Roles:    roles,
		ShowHelp: help,
	})
}

func (srv *Roles) get(w http.ResponseWriter, r *http.Request) {
	roleID, _ := strconv.ParseInt(r.PathValue("roleID"), 10, 64)
	role, err := srv.dreamer.DB().GetRole(r.Context(), roleID)
	if err != nil {
		views.RenderError(w, err)
		return
	}
	isEdit, _ := strconv.ParseBool(r.FormValue("edit"))
	if isEdit {
		viewEdit().HTML(w, editParams{
			Role:     role,
			Purposes: common.PurposeValues(),
		})
	} else {
		viewIndex().HTML(w, indexParams{
			Role: role,
		})
	}
}

func (srv *Roles) wizard(w http.ResponseWriter, r *http.Request) {
	viewNew().HTML(w, newParams{
		Purposes:    common.PurposeValues(),
		Content:     r.FormValue("content"),
		Description: r.FormValue("description"),
		Model:       r.FormValue("model"),
		Purpose:     common.Purpose(r.FormValue("purpose")),
	})
}

type roleParams struct {
	Name    string         `schema:"name"`
	System  string         `schema:"system"`
	Model   string         `schema:"model"`
	Purpose common.Purpose `schema:"purpose"`
}

func (srv *Roles) create(w http.ResponseWriter, r *http.Request) {
	params, err := views.BindForm[roleParams](r)
	if err != nil {
		views.RenderError(w, err)
		return
	}

	role, err := srv.dreamer.DB().CreateRole(r.Context(), dbo.CreateRoleParams{
		Name:    params.Name,
		System:  params.System,
		Model:   params.Model,
		Purpose: params.Purpose,
	})
	if err != nil {
		views.RenderError(w, err)
		return
	}
	w.Header().Set("Location", strconv.FormatInt(role.ID, 10)+"/")
	w.WriteHeader(http.StatusSeeOther)
}

func (srv *Roles) update(w http.ResponseWriter, r *http.Request) {
	params, err := views.BindForm[roleParams](r)
	if err != nil {
		views.RenderError(w, err)
		return
	}
	roleID, _ := strconv.ParseInt(r.PathValue("roleID"), 10, 64)
	_, err = srv.dreamer.DB().UpdateRole(r.Context(), dbo.UpdateRoleParams{
		Name:    params.Name,
		System:  params.System,
		Model:   params.Model,
		Purpose: params.Purpose,
		ID:      roleID,
	})
	if err != nil {
		views.RenderError(w, err)
		return
	}
	w.Header().Set("Location", ".?"+r.URL.RawQuery)
	w.WriteHeader(http.StatusSeeOther)
}

func (srv *Roles) delete(w http.ResponseWriter, r *http.Request) {
	roleID, _ := strconv.ParseInt(r.PathValue("roleID"), 10, 64)
	if err := srv.dreamer.DB().DeleteRole(r.Context(), roleID); err != nil {
		views.RenderError(w, err)
		return
	}
	if views.IsHTMX(r) {
		if r.FormValue("stay") == "true" {
			w.WriteHeader(http.StatusOK)
			return
		}
		w.Header().Set("HX-Redirect", "../")
	} else {
		w.Header().Set("Location", "../")
	}
	w.WriteHeader(http.StatusSeeOther)
}

type meta struct {
	Title   string         `yaml:"title,omitempty"`
	Model   string         `yaml:"model,omitempty"`
	Purpose common.Purpose `yaml:"purpose"`
}

func (srv *Roles) exportOne(w http.ResponseWriter, r *http.Request) {
	roleID, _ := strconv.ParseInt(r.PathValue("roleID"), 10, 64)
	role, err := srv.dreamer.DB().GetRole(r.Context(), roleID)
	if err != nil {
		views.RenderError(w, err)
		return
	}

	content, err := mark.Build(meta{
		Title:   role.Name,
		Model:   role.Model,
		Purpose: role.Purpose,
	}, role.System)
	if err != nil {
		views.RenderError(w, err)
		return
	}
	name := strings.TrimSpace(path.Base(role.Name))
	if name == "" {
		name = strconv.FormatInt(role.ID, 10)
	}
	w.Header().Set("Content-Type", "text/markdown; charset=utf-8")
	w.Header().Set("Content-Disposition", "attachment; filename="+strconv.Quote(name+".md"))
	_, _ = w.Write([]byte(content))
}

func (srv *Roles) exportAll(w http.ResponseWriter, r *http.Request) {
	roles, err := srv.dreamer.DB().ListRoles(r.Context())
	if err != nil {
		views.RenderError(w, err)
		return
	}
	var maxID int64
	for _, role := range roles {
		maxID = max(maxID, role.ID)
	}
	digits := len(strconv.FormatInt(maxID, 10))
	format := "%0" + strconv.Itoa(digits) + "d-%s.md"
	w.Header().Set("Content-Type", "application/zip")
	w.Header().Set("Content-Disposition", "attachment; filename=roles.zip")
	out := zip.NewWriter(w)
	defer out.Close()
	for _, role := range roles {
		content, err := mark.Build(meta{
			Title:   role.Name,
			Model:   role.Model,
			Purpose: role.Purpose,
		}, role.System)
		if err != nil {
			views.RenderError(w, err)
			return
		}

		name := fmt.Sprintf(format, role.ID, strings.TrimSpace(path.Base(role.Name)))

		f, err := out.Create(name)
		if err != nil {
			views.RenderError(w, err)
			return
		}
		_, err = f.Write([]byte(content))
		if err != nil {
			views.RenderError(w, err)
			return
		}
	}
}

func (srv *Roles) importWizard(w http.ResponseWriter, r *http.Request) {
	viewImport().HTML(w, importParams{})
}

func (srv *Roles) importAll(w http.ResponseWriter, r *http.Request) {
	err := lifecycle.Import[meta](r, func(d mark.Document[meta], name string) error {
		if d.Metadata.Title != "" {
			name = d.Metadata.Title
		}
		if d.Metadata.Purpose == "" {
			d.Metadata.Purpose = common.PurposeWrite
		}

		_, err := srv.dreamer.DB().CreateRole(r.Context(), dbo.CreateRoleParams{
			Name:    name,
			System:  d.Content,
			Model:   d.Metadata.Model,
			Purpose: d.Metadata.Purpose,
		})
		return err
	})
	if err != nil {
		views.RenderError(w, err)
		return
	}

	w.Header().Set("Location", ".")
	w.WriteHeader(http.StatusSeeOther)
}
