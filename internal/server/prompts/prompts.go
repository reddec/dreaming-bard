package prompts

import (
	"archive/zip"
	"context"
	"fmt"
	"net/http"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/reddec/dreaming-bard/internal/common"
	"github.com/reddec/dreaming-bard/internal/dbo"
	"github.com/reddec/dreaming-bard/internal/dreamwriter"
	"github.com/reddec/dreaming-bard/internal/server/views"
	"github.com/reddec/dreaming-bard/internal/utils/lifecycle"
	"github.com/reddec/dreaming-bard/internal/utils/mark"
)

func New(dreamer *dreamwriter.DreamWriter) *Prompts {
	mux := http.NewServeMux()
	srv := &Prompts{
		Handler:  mux,
		dreamer:  dreamer,
		showHelp: dbo.NewPref[bool](dreamer.DB(), "help_prompt", true),
	}
	mux.HandleFunc("GET /", srv.list)
	mux.HandleFunc("POST /", srv.create)
	mux.HandleFunc("GET /new", srv.wizard)
	mux.HandleFunc("GET /import", srv.importWizard)
	mux.HandleFunc("GET /export", srv.exportAll)
	mux.HandleFunc("POST /import", srv.importAll)
	mux.HandleFunc("POST /help", views.BoolHandler(srv.showHelp))
	mux.HandleFunc("GET /{promptID}/{$}", srv.edit)
	mux.HandleFunc("POST /{promptID}/{$}", srv.update)
	mux.HandleFunc("DELETE /{promptID}/", srv.delete)
	mux.HandleFunc("POST /{promptID}/pin", srv.setPin)
	return srv
}

type Prompts struct {
	http.Handler
	dreamer  *dreamwriter.DreamWriter
	showHelp *dbo.Pref[bool]
}

func (srv *Prompts) Run(ctx context.Context) error {
	// TODO: background tasks
	return nil
}

func (srv *Prompts) list(w http.ResponseWriter, r *http.Request) {
	prompts, err := srv.dreamer.DB().ListPrompts(r.Context())
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
		Prompts:  prompts,
		ShowHelp: help,
	})
}

func (srv *Prompts) wizard(w http.ResponseWriter, r *http.Request) {
	roles, err := srv.dreamer.DB().ListRoles(r.Context())
	if err != nil {
		views.RenderError(w, err)
		return
	}
	viewNew().HTML(w, newParams{
		Roles: roles,
	})
}

type promptForm struct {
	Summary     string
	Content     string
	DefaultRole int64 `schema:"default_role"`
}

func (srv *Prompts) create(w http.ResponseWriter, r *http.Request) {
	params, err := views.BindForm[promptForm](r)
	if err != nil {
		views.RenderError(w, err)
		return
	}
	p, err := srv.dreamer.DB().CreatePrompt(r.Context(), dbo.CreatePromptParams{
		Summary: params.Summary,
		Content: params.Content,
		RoleID:  params.DefaultRole,
	})
	if err != nil {
		views.RenderError(w, err)
		return
	}
	w.Header().Set("Location", "./"+strconv.FormatInt(p.ID, 10)+"/")
	w.WriteHeader(http.StatusSeeOther)
}

func (srv *Prompts) setPin(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(r.PathValue("promptID"), 10, 64)
	pin, _ := strconv.ParseBool(r.FormValue("pin"))
	var pinnedAt *time.Time
	if pin {
		v := time.Now()
		pinnedAt = &v
	}
	err := srv.dreamer.DB().UpdatePromptPin(r.Context(), dbo.UpdatePromptPinParams{
		PinnedAt: pinnedAt,
		ID:       id,
	})
	if err != nil {
		views.RenderError(w, err)
		return
	}
	w.Header().Set("Location", "../")
	w.WriteHeader(http.StatusSeeOther)
}

func (srv *Prompts) edit(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(r.PathValue("promptID"), 10, 64)
	prompt, err := srv.dreamer.DB().GetPrompt(r.Context(), id)
	if err != nil {
		views.RenderError(w, err)
		return
	}
	roles, err := srv.dreamer.DB().ListRoles(r.Context())
	if err != nil {
		views.RenderError(w, err)
		return
	}
	viewEdit().HTML(w, editParams{
		Roles:  roles,
		Prompt: prompt,
	})
}

func (srv *Prompts) update(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(r.PathValue("promptID"), 10, 64)
	params, err := views.BindForm[promptForm](r)
	if err != nil {
		views.RenderError(w, err)
		return
	}
	err = srv.dreamer.DB().UpdatePrompt(r.Context(), dbo.UpdatePromptParams{
		Summary: params.Summary,
		Content: params.Content,
		RoleID:  params.DefaultRole,
		ID:      id,
	})
	if err != nil {
		views.RenderError(w, err)
		return
	}
	w.Header().Set("Location", "./")
	w.WriteHeader(http.StatusSeeOther)
}

func (srv *Prompts) delete(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(r.PathValue("promptID"), 10, 64)
	if err := srv.dreamer.DB().DeletePrompt(r.Context(), id); err != nil {
		views.RenderError(w, err)
		return
	}
	if views.IsHTMX(r) {
		if r.FormValue("stay") != "true" {
			w.Header().Set("HX-Redirect", "../")
		}
		w.WriteHeader(http.StatusOK)
	} else {
		w.Header().Set("Location", "../")
		w.WriteHeader(http.StatusSeeOther)
	}
}

func (srv *Prompts) importWizard(w http.ResponseWriter, r *http.Request) {
	roles, err := srv.dreamer.DB().ListRoles(r.Context())
	if err != nil {
		views.RenderError(w, err)
		return
	}
	var firstWriter int64
	for _, role := range roles {
		if role.Purpose == common.PurposeWrite {
			firstWriter = role.ID
			break
		}
	}
	viewImport().HTML(w, importParams{
		Roles:       roles,
		FirstWriter: firstWriter,
	})
}

type meta struct {
	Summary string `yaml:"summary,omitempty"`
	Role    string `yaml:"role,omitempty"`
}

func (srv *Prompts) exportAll(w http.ResponseWriter, r *http.Request) {
	prompts, err := srv.dreamer.DB().ListPrompts(r.Context())
	if err != nil {
		views.RenderError(w, err)
		return
	}

	var maxID int64
	for _, prompt := range prompts {
		maxID = max(maxID, prompt.Prompt.ID)
	}
	digits := len(strconv.FormatInt(maxID, 10))
	format := "%0" + strconv.Itoa(digits) + "d-%s.md"
	w.Header().Set("Content-Type", "application/zip")
	w.Header().Set("Content-Disposition", "attachment; filename=prompts.zip")
	out := zip.NewWriter(w)
	defer out.Close()
	for _, prompt := range prompts {
		content, err := mark.Build(meta{
			Summary: prompt.Prompt.Summary,
			Role:    prompt.RoleName,
		}, prompt.Prompt.Content)
		if err != nil {
			views.RenderError(w, err)
			return
		}

		name := fmt.Sprintf(format, prompt.Prompt.ID, strings.TrimSpace(path.Base(strings.Split(strings.TrimSpace(prompt.Prompt.Summary), "\n")[0])))

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

func (srv *Prompts) importAll(w http.ResponseWriter, r *http.Request) {
	roles, err := srv.dreamer.DB().ListRoles(r.Context())
	if err != nil {
		views.RenderError(w, err)
		return
	}
	var roleMap = make(map[string]*dbo.Role, len(roles))
	for _, role := range roles {
		roleMap[role.Name] = &role
	}

	if len(roleMap) == 0 {
		views.RenderError(w, fmt.Errorf("no roles found, please create a role first"))
		return
	}

	defaultRoleID, _ := strconv.ParseInt(r.FormValue("default_role"), 10, 64)
	if defaultRoleID <= 0 {
		// first writer is default role
		for _, role := range roles {
			if role.Purpose == common.PurposeWrite {
				defaultRoleID = role.ID
				break
			}
		}
		// oops - no writer, use at least something
		if defaultRoleID == 0 {
			defaultRoleID = roles[0].ID
		}
	}

	err = lifecycle.Import[meta](r, func(d mark.Document[meta], summary string) error {
		if d.Metadata.Summary != "" {
			summary = d.Metadata.Summary
		}

		// map by name
		var roleID int64
		if d.Metadata.Role != "" {
			if role, ok := roleMap[d.Metadata.Role]; ok {
				roleID = role.ID
			} else {
				roleID = defaultRoleID
			}
		}

		_, err := srv.dreamer.DB().CreatePrompt(r.Context(), dbo.CreatePromptParams{
			Summary: summary,
			Content: d.Content,
			RoleID:  roleID,
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
