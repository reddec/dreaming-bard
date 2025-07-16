package blueprint

import (
	"context"
	"encoding/xml"
	"fmt"
	"log/slog"
	"net/http"
	"slices"
	"strconv"
	"strings"

	"github.com/sourcegraph/conc/pool"

	"github.com/reddec/dreaming-bard/internal/dbo"
	"github.com/reddec/dreaming-bard/internal/dreamwriter"
	"github.com/reddec/dreaming-bard/internal/server/views"
	"github.com/reddec/dreaming-bard/internal/utils/xsync"
)

type BackgroundChat interface {
	Submit(chatID int64, options ...dreamwriter.RunOption) error
}

func New(dreamer *dreamwriter.DreamWriter, chatter BackgroundChat) *Blueprint {
	mux := http.NewServeMux()
	bp := &Blueprint{
		Handler:  mux,
		dreamer:  dreamer,
		chatter:  chatter,
		showHelp: dbo.NewPref[bool](dreamer.DB(), "help_blueprint", true),
		enhancer: xsync.NewPool[dbo.BlueprintStep](),
		planner:  xsync.NewPool[dbo.Blueprint](),
	}

	mux.HandleFunc("GET /", bp.list)
	mux.HandleFunc("GET /new", bp.wizard)
	mux.HandleFunc("POST /help", views.BoolHandler(bp.showHelp))
	mux.HandleFunc("GET /{id}/", bp.index)
	mux.HandleFunc("POST /{$}", bp.create)
	mux.HandleFunc("POST /{id}/", bp.update)
	mux.HandleFunc("POST /{id}/plan", bp.plan)
	mux.HandleFunc("POST /{id}/chat", bp.startChat)
	mux.HandleFunc("POST /{id}/steps", bp.createStep)
	mux.HandleFunc("POST /{id}/steps/{stepID}", bp.updateStep)
	mux.HandleFunc("POST /{id}/steps/{stepID}/enhance", bp.enhance)
	mux.HandleFunc("DELETE /{id}/steps/{stepID}", bp.deleteStep)

	mux.HandleFunc("POST /{id}/contexts", bp.linkContext)
	mux.HandleFunc("DELETE /{id}/contexts/{contextID}", bp.unlinkContext)

	mux.HandleFunc("POST /{id}/pages", bp.setPages)
	return bp
}

type Blueprint struct {
	http.Handler
	showHelp *dbo.Pref[bool]
	dreamer  *dreamwriter.DreamWriter
	chatter  BackgroundChat
	enhancer *xsync.Pool[dbo.BlueprintStep]
	planner  *xsync.Pool[dbo.Blueprint]
}

func (bp *Blueprint) Run(ctx context.Context) error {
	p := pool.New().WithContext(ctx).WithCancelOnError()
	p.Go(func(ctx context.Context) error {
		slog.Info("starting blueprint enhancer")
		bp.enhancer.Run(ctx)
		return nil
	})
	p.Go(func(ctx context.Context) error {
		slog.Info("starting blueprint planner")
		bp.planner.Run(ctx)
		return nil
	})
	return p.Wait()
}

func (bp *Blueprint) list(w http.ResponseWriter, r *http.Request) {
	list, err := bp.dreamer.DB().ListBlueprints(r.Context())
	if err != nil {
		views.RenderError(w, err)
		return
	}
	help, err := bp.showHelp.Get(r.Context())
	if err != nil {
		views.RenderError(w, err)
		return
	}
	viewList().HTML(w, listParams{Blueprints: list, ShowHelp: help})
}

func (bp *Blueprint) index(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(r.PathValue("id"), 10, 64)
	blueprint, err := bp.dreamer.DB().GetBlueprint(r.Context(), id)
	if err != nil {
		views.RenderError(w, err)
		return
	}
	steps, err := bp.dreamer.DB().ListBlueprintSteps(r.Context(), id)
	if err != nil {
		views.RenderError(w, err)
		return
	}
	availableContext, err := bp.dreamer.DB().ListBlueprintUnlinkedContexts(r.Context(), id)
	if err != nil {
		views.RenderError(w, err)
		return
	}

	linkedContext, err := bp.dreamer.DB().ListBlueprintLinkedContexts(r.Context(), id)
	if err != nil {
		views.RenderError(w, err)
		return
	}

	linkedChats, err := bp.dreamer.DB().ListBlueprintChats(r.Context(), id)
	if err != nil {
		views.RenderError(w, err)
		return
	}

	pages, err := bp.dreamer.DB().ListBlueprintPages(r.Context(), id)
	if err != nil {
		views.RenderError(w, err)
		return
	}

	roles, err := bp.dreamer.DB().ListRoles(r.Context())
	if err != nil {
		views.RenderError(w, err)
		return
	}

	editStep, _ := strconv.ParseInt(r.FormValue("editStep"), 10, 64)

	viewIndex().HTML(w, indexParams{
		Blueprint:        blueprint,
		Pages:            pages,
		Steps:            steps,
		EditStep:         editStep,
		AvailableContext: availableContext,
		LinkedContext:    linkedContext,
		Roles:            roles,
		Chats:            linkedChats,
		Tasks:            bp.enhancer.List(),
		PlanningTasks:    bp.planner.List(),
	})
}

func (bp *Blueprint) create(w http.ResponseWriter, r *http.Request) {
	item, err := bp.dreamer.DB().CreateBlueprint(r.Context())
	if err != nil {
		views.RenderError(w, err)
		return
	}
	if views.IsHTMX(r) {
		w.Header().Set("Hx-Redirect", "./"+strconv.FormatInt(item.ID, 10)+"/")
	} else {
		w.Header().Set("Location", strconv.FormatInt(item.ID, 10)+"/")
		w.WriteHeader(http.StatusSeeOther)
	}
}

func (bp *Blueprint) wizard(w http.ResponseWriter, r *http.Request) {
	pages, err := bp.dreamer.DB().ListPages(r.Context())
	if err != nil {
		views.RenderError(w, err)
		return
	}
	viewWizard().HTML(w, wizardParams{Pages: pages})
}

func (bp *Blueprint) update(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(r.PathValue("id"), 10, 64)

	form, err := views.BindForm[struct {
		Note string `schema:"note"`
	}](r)
	if err != nil {
		views.RenderError(w, err)
		return
	}

	err = bp.dreamer.DB().UpdateBlueprint(r.Context(), dbo.UpdateBlueprintParams{
		Note: form.Note,
		ID:   id,
	})
	if err != nil {
		views.RenderError(w, err)
		return
	}
	if views.IsHTMX(r) {
		w.WriteHeader(http.StatusOK)
	} else {
		w.Header().Set("Location", "./")
		w.WriteHeader(http.StatusSeeOther)
	}
}

func (bp *Blueprint) createStep(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(r.PathValue("id"), 10, 64)

	_, err := bp.dreamer.DB().CreateBlueprintStep(r.Context(), dbo.CreateBlueprintStepParams{
		BlueprintID: id,
		Content:     r.FormValue("content"),
	})
	if err != nil {
		views.RenderError(w, err)
		return
	}
	w.Header().Set("Location", "./")
	w.WriteHeader(http.StatusSeeOther)
}

func (bp *Blueprint) deleteStep(w http.ResponseWriter, r *http.Request) {
	// id, _ := strconv.ParseInt(r.PathValue("id"), 10, 64)
	stepID, _ := strconv.ParseInt(r.PathValue("stepID"), 10, 64)
	err := bp.dreamer.DB().DeleteBlueprintStep(r.Context(), stepID)
	if err != nil {
		views.RenderError(w, err)
		return
	}
	if views.IsHTMX(r) {
		w.WriteHeader(http.StatusOK)
	} else {
		w.Header().Set("Location", "../")
		w.WriteHeader(http.StatusSeeOther)
	}
}

func (bp *Blueprint) updateStep(w http.ResponseWriter, r *http.Request) {
	stepID, _ := strconv.ParseInt(r.PathValue("stepID"), 10, 64)
	err := bp.dreamer.DB().UpdateBlueprintStep(r.Context(), dbo.UpdateBlueprintStepParams{
		Content: r.FormValue("content"),
		ID:      stepID,
	})
	if err != nil {
		views.RenderError(w, err)
		return
	}
	w.Header().Set("Location", "../")
	w.WriteHeader(http.StatusSeeOther)
}

func (bp *Blueprint) linkContext(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(r.PathValue("id"), 10, 64)
	contextID, _ := strconv.ParseInt(r.FormValue("contextID"), 10, 64)
	err := bp.dreamer.DB().BlueprintLinkContext(r.Context(), dbo.BlueprintLinkContextParams{
		BlueprintID: id,
		ContextID:   contextID,
	})
	if err != nil {
		views.RenderError(w, err)
		return
	}
	w.Header().Set("Location", "./")
	w.WriteHeader(http.StatusSeeOther)
}

func (bp *Blueprint) unlinkContext(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(r.PathValue("id"), 10, 64)
	contextID, _ := strconv.ParseInt(r.PathValue("contextID"), 10, 64)
	err := bp.dreamer.DB().BlueprintUnlinkContext(r.Context(), dbo.BlueprintUnlinkContextParams{
		BlueprintID: id,
		ContextID:   contextID,
	})
	if err != nil {
		views.RenderError(w, err)
		return
	}
	if views.IsHTMX(r) {
		w.WriteHeader(http.StatusOK)
	} else {
		w.Header().Set("Location", "../")
		w.WriteHeader(http.StatusSeeOther)

	}
}

func (bp *Blueprint) setPages(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(r.PathValue("id"), 10, 64)

	if err := r.ParseForm(); err != nil {
		views.RenderError(w, err)
		return
	}

	for key, values := range r.Form {
		if len(values) == 0 {
			continue
		}
		pageID, err := strconv.ParseInt(key, 10, 64)
		if err != nil {
			continue
		}
		switch values[0] {
		case "full":
			if err := bp.dreamer.DB().SetBlueprintLinkedPage(r.Context(), dbo.SetBlueprintLinkedPageParams{
				BlueprintID: id,
				PageID:      pageID,
				Inline:      true,
			}); err != nil {
				views.RenderError(w, err)
				return
			}
		case "summary":
			if err := bp.dreamer.DB().SetBlueprintLinkedPage(r.Context(), dbo.SetBlueprintLinkedPageParams{
				BlueprintID: id,
				PageID:      pageID,
				Inline:      false,
			}); err != nil {
				views.RenderError(w, err)
				return
			}
		case "ignore":
			if err := bp.dreamer.DB().BlueprintUnlinkPage(r.Context(), dbo.BlueprintUnlinkPageParams{
				BlueprintID: id,
				PageID:      pageID,
			}); err != nil {
				views.RenderError(w, err)
				return
			}
		}
	}

	w.Header().Set("Location", "./")
	w.WriteHeader(http.StatusSeeOther)
}

func (bp *Blueprint) startChat(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(r.PathValue("id"), 10, 64)
	roleID, _ := strconv.ParseInt(r.FormValue("roleID"), 10, 64)
	chat, err := bp.dreamer.NewChat(r.Context(), roleID)
	if err != nil {
		views.RenderError(w, err)
		return
	}
	err = bp.dreamer.DB().LinkBlueprintChat(r.Context(), dbo.LinkBlueprintChatParams{
		BlueprintID: id,
		ChatID:      chat.Entity().ID,
	})
	if err != nil {
		views.RenderError(w, err)
		return
	}

	_, err = bp.addContext(r.Context(), id, chat)
	if err != nil {
		views.RenderError(w, err)
		return
	}

	steps, err := bp.dreamer.DB().ListBlueprintSteps(r.Context(), id)
	if err != nil {
		views.RenderError(w, err)
		return
	}

	message, err := buildUserOutline(steps)
	if err != nil {
		views.RenderError(w, err)
		return
	}

	if err := chat.User(r.Context(), message); err != nil {
		views.RenderError(w, err)
		return
	}
	err = bp.chatter.Submit(chat.Entity().ID)
	if err != nil {
		views.RenderError(w, err)
		return
	}
	w.Header().Set("Location", "/chats/"+strconv.FormatInt(chat.Entity().ID, 10)+"/#end")
	w.WriteHeader(http.StatusSeeOther)
}

func (bp *Blueprint) addContext(ctx context.Context, id int64, chat *dreamwriter.Chat) (*dbo.Blueprint, error) {
	linkedContext, err := bp.dreamer.DB().ListBlueprintLinkedContexts(ctx, id)
	if err != nil {
		return nil, err
	}
	linkedPages, err := bp.dreamer.DB().ListBlueprintLinkedPages(ctx, id)
	if err != nil {
		return nil, err
	}
	plan, err := bp.dreamer.DB().GetBlueprint(ctx, id)
	if err != nil {
		return nil, err
	}

	for _, doc := range linkedContext {
		if err := chat.AddContexts(ctx, doc.ID); err != nil {
			return nil, fmt.Errorf("add context %d to chat: %w", doc.ID, err)
		}
	}

	// chronological order
	slices.SortFunc(linkedPages, func(a, b dbo.ListBlueprintLinkedPagesRow) int {
		return int(a.Page.Num - b.Page.Num)
	})

	for _, doc := range linkedPages {
		if err := chat.AddPage(ctx, doc.Page.ID, doc.Inline); err != nil {
			return nil, fmt.Errorf("add page %d to chat: %w", doc.Page.ID, err)
		}
	}

	if plan.Note != "" {
		if err := chat.AddNote(ctx, plan.Note); err != nil {
			return nil, fmt.Errorf("add note to chat: %w", err)
		}
	}

	return &plan, nil
}

func (bp *Blueprint) enhance(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(r.PathValue("id"), 10, 64)
	roleID, _ := strconv.ParseInt(r.FormValue("roleID"), 10, 64)
	stepID, _ := strconv.ParseInt(r.PathValue("stepID"), 10, 64)
	step, err := bp.dreamer.DB().GetBlueprintStep(r.Context(), stepID)
	if err != nil {
		views.RenderError(w, err)
		return
	}

	chat, err := bp.dreamer.NewChat(r.Context(), roleID, dreamwriter.Annotation("enhance step"))
	if err != nil {
		views.RenderError(w, err)
		return
	}

	_, err = bp.addContext(r.Context(), step.BlueprintID, chat)
	if err != nil {
		views.RenderError(w, err)
		return
	}

	previousSteps, err := bp.dreamer.DB().ListBlueprintPreviousSteps(r.Context(), dbo.ListBlueprintPreviousStepsParams{
		BlueprintID: id,
		ID:          stepID,
	})
	if err != nil {
		views.RenderError(w, err)
		return
	}
	var previously string
	if len(previousSteps) > 0 {
		for _, s := range previousSteps {
			previously += "\n\n" + s.Content
		}
		previously = strings.TrimSpace(previously)

		if err := chat.AddDocument(r.Context(), " aggregated information about previous beats", previously); err != nil {
			views.RenderError(w, err)
			return
		}
	}

	if err := chat.User(r.Context(), step.Content); err != nil {
		views.RenderError(w, err)
		return
	}

	err = bp.enhancer.Try(step, func(ctx context.Context) {
		out, err := chat.Run(ctx)
		if err != nil {
			slog.Error("failed to run shot", "error", err)
			return
		}
		err = bp.dreamer.DB().UpdateBlueprintStep(ctx, dbo.UpdateBlueprintStepParams{
			Content: out,
			ID:      step.ID,
		})
		if err != nil {
			slog.Error("failed to update step", "error", err)
		}
	})
	if err != nil {
		views.RenderError(w, err)
		return
	}
	if views.IsHTMX(r) {
		w.Header().Set("HX-Redirect", "./#s"+strconv.FormatInt(step.ID, 10))
	} else {
		w.Header().Set("Location", "../../#s"+strconv.FormatInt(step.ID, 10))
	}
	w.WriteHeader(http.StatusSeeOther)
}

func (bp *Blueprint) plan(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(r.PathValue("id"), 10, 64)
	roleID, _ := strconv.ParseInt(r.FormValue("roleID"), 10, 64)

	chat, err := bp.dreamer.NewChat(r.Context(), roleID, dreamwriter.Annotation("plan blueprint"))
	if err != nil {
		views.RenderError(w, err)
		return
	}

	info, err := bp.addContext(r.Context(), id, chat)
	if err != nil {
		views.RenderError(w, err)
		return
	}

	if err := chat.User(r.Context(), r.FormValue("content")); err != nil {
		views.RenderError(w, err)
		return
	}

	err = bp.planner.Try(*info, func(ctx context.Context) {
		out, err := chat.Run(ctx)
		if err != nil {
			slog.Error("failed to run shot", "error", err)
			return
		}

		steps := strings.Split(out, "---")
		for _, step := range steps {
			step = strings.TrimSpace(step)
			if step == "" {
				continue
			}
			_, err = bp.dreamer.DB().CreateBlueprintStep(ctx, dbo.CreateBlueprintStepParams{
				BlueprintID: id,
				Content:     step,
			})
			if err != nil {
				slog.Error("failed to create step", "error", err)
			}
		}

		if err != nil {
			slog.Error("failed to update step", "error", err)
		}
	})
	if err != nil {
		views.RenderError(w, err)
		return
	}
	if views.IsHTMX(r) {
		w.Header().Set("HX-Redirect", "./")
	} else {
		w.Header().Set("Location", "./")
	}
	w.WriteHeader(http.StatusSeeOther)
}

type Beat struct {
	XMLName xml.Name `xml:"beat"`
	Content string   `xml:",chardata"`
}

func buildUserOutline(steps []dbo.BlueprintStep) (string, error) {
	var outline []Beat
	for _, step := range steps {
		if step.Content == "" {
			continue
		}
		outline = append(outline, Beat{Content: strings.TrimSpace(step.Content)})
	}
	v, err := xml.MarshalIndent(outline, "", "  ")
	return "<outline>\n" + string(v) + "\n</outline>", err
}
