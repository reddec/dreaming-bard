package chats

import (
	"context"
	"log/slog"
	"net/http"
	"slices"
	"strconv"
	"strings"

	"github.com/reddec/dreaming-bard/internal/dbo"
	"github.com/reddec/dreaming-bard/internal/dreamwriter"
	"github.com/reddec/dreaming-bard/internal/server/views"
	"github.com/reddec/dreaming-bard/internal/utils/xsync"
)

func New(dreamer *dreamwriter.DreamWriter) *Chats {
	mux := http.NewServeMux()
	srv := &Chats{
		Handler:  mux,
		dreamer:  dreamer,
		worker:   xsync.NewPool[int64](),
		showHelp: dbo.NewPref[bool](dreamer.DB(), "help_chats", true),
	}
	mux.HandleFunc("GET /{$}", srv.list)
	mux.HandleFunc("GET /new", srv.wizard)
	mux.HandleFunc("POST /{$}", srv.create)
	mux.HandleFunc("POST /help", views.BoolHandler(srv.showHelp))
	mux.HandleFunc("GET /{threadID}/", srv.index)
	mux.HandleFunc("POST /{threadID}/", srv.send)
	mux.HandleFunc("POST /{threadID}/stop", srv.stop)
	mux.HandleFunc("POST /{threadID}/draft", srv.saveDraft)
	mux.HandleFunc("DELETE /{threadID}/messages/{messageID}/", srv.deleteMessage)
	mux.HandleFunc("POST /{threadID}/messages/{messageID}/", srv.updateMessageContent)

	return srv
}

type Chats struct {
	http.Handler
	dreamer  *dreamwriter.DreamWriter
	worker   *xsync.Pool[int64]
	showHelp *dbo.Pref[bool]
}

func (srv *Chats) Run(ctx context.Context) error {
	srv.worker.Run(ctx)
	return nil
}

func (srv *Chats) Submit(chatID int64, options ...dreamwriter.RunOption) error {
	// FIXME: it should not be responsibility of a controller - move out

	return srv.worker.Try(chatID, func(ctx context.Context) {
		chat, err := srv.dreamer.OpenChat(ctx, chatID)
		if err != nil {
			slog.Error("failed to open chat", "chat", chatID, "error", err)
			return
		}
		_, err = chat.Run(ctx, options...)
		if err != nil {
			slog.Error("failed to run chat", "chat", chat.Entity().ID, "error", err)
		}
	})
}

func (srv *Chats) list(w http.ResponseWriter, r *http.Request) {
	showHelp, err := srv.showHelp.Get(r.Context())
	if err != nil {
		views.RenderError(w, err)
		return
	}
	threads, err := srv.dreamer.DB().ListChats(r.Context())
	if err != nil {
		views.RenderError(w, err)
		return
	}

	viewList().HTML(w, listParams{
		Threads:   threads,
		BusyChats: srv.worker.RunningStates(),
		ShowHelp:  showHelp,
	})
}

func (srv *Chats) wizard(w http.ResponseWriter, r *http.Request) {
	v, _ := strconv.ParseInt(r.FormValue("promptID"), 10, 64)

	facts, err := srv.dreamer.DB().ListContexts(r.Context())
	if err != nil {
		views.RenderError(w, err)
		return
	}
	roles, err := srv.dreamer.DB().ListRoles(r.Context())
	if err != nil {
		views.RenderError(w, err)
		return
	}

	var prompt dbo.Prompt
	if v != 0 {
		prompt, err = srv.dreamer.DB().GetPrompt(r.Context(), v)
	}
	if err != nil {
		views.RenderError(w, err)
		return
	}

	viewWizard().HTML(w, wizardParams{
		Prompt: prompt,
		Facts:  facts,
		Roles:  roles,
	})
}

func (srv *Chats) create(w http.ResponseWriter, r *http.Request) {
	role, _ := strconv.ParseInt(r.FormValue("role"), 10, 64)

	thread, err := srv.dreamer.NewChat(r.Context(), role,
		dreamwriter.Draft(r.FormValue("draft")),
		dreamwriter.Annotation("user chat"),
	)
	if err != nil {
		views.RenderError(w, err)
		return
	}
	w.Header().Set("Location", strconv.FormatInt(thread.Entity().ID, 10)+"/")
	w.WriteHeader(http.StatusSeeOther)
}

func (srv *Chats) index(w http.ResponseWriter, r *http.Request) {
	threadID, _ := strconv.ParseInt(r.PathValue("threadID"), 10, 64)
	thread, err := srv.dreamer.DB().GetChat(r.Context(), threadID)
	if err != nil {
		views.RenderError(w, err)
		return
	}
	messages, err := srv.dreamer.DB().ListMessagesByChat(r.Context(), threadID)
	if err != nil {
		views.RenderError(w, err)
		return
	}

	contexts, err := srv.dreamer.DB().ListContexts(r.Context())
	if err != nil {
		views.RenderError(w, err)
		return
	}

	role, err := srv.dreamer.DB().GetRole(r.Context(), thread.RoleID)
	if err != nil {
		views.RenderError(w, err)
		return
	}

	pages, err := srv.dreamer.DB().ListPages(r.Context())
	if err != nil {
		views.RenderError(w, err)
		return
	}
	slices.SortFunc(pages, func(a, b dbo.Page) int {
		return int(b.Num) - int(a.Num)
	})

	editMessageID, _ := strconv.ParseInt(r.FormValue("editMessageID"), 10, 64)

	var isBusy bool
	for _, active := range srv.worker.List() {
		if active.State() == threadID {
			isBusy = active.IsRunning()
			break
		}
	}

	viewIndex().HTML(w, indexParams{
		Role:          role,
		Chat:          thread,
		History:       messages,
		UID:           threadID,
		IsBusy:        isBusy,
		Facts:         contexts,
		EditMessageID: editMessageID,
		Pages:         pages,
	})
}

func (srv *Chats) stop(w http.ResponseWriter, r *http.Request) {
	threadID, _ := strconv.ParseInt(r.PathValue("threadID"), 10, 64)

	for _, active := range srv.worker.List() {
		if active.State() == threadID {
			active.Stop(r.Context())
			break
		}
	}

	w.Header().Set("Location", ".#end")
	w.WriteHeader(http.StatusSeeOther)
}

func (srv *Chats) saveDraft(w http.ResponseWriter, r *http.Request) {
	threadID, _ := strconv.ParseInt(r.PathValue("threadID"), 10, 64)
	if err := srv.dreamer.DB().UpdateChatDraft(r.Context(), dbo.UpdateChatDraftParams{
		Draft: r.FormValue("message"),
		ID:    threadID,
	}); err != nil {
		views.RenderError(w, err)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (srv *Chats) send(w http.ResponseWriter, r *http.Request) {
	threadID, _ := strconv.ParseInt(r.PathValue("threadID"), 10, 64)
	message := strings.TrimSpace(r.FormValue("message"))
	var contextIDs []int64
	for _, v := range r.PostForm["fact"] {
		id, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			views.RenderError(w, err)
			return
		}
		contextIDs = append(contextIDs, id)
	}

	thread, err := srv.dreamer.OpenChat(r.Context(), threadID)
	if err != nil {
		views.RenderError(w, err)
		return
	}

	if err := thread.AddContexts(r.Context(), contextIDs...); err != nil {
		views.RenderError(w, err)
		return
	}

	type pageRef struct {
		ID   int64
		Full bool
		Num  int64
	}
	var pagesToUse []pageRef
	for fieldName, values := range r.PostForm {
		if len(values) == 0 {
			continue
		}
		kind, idStr, ok := strings.Cut(fieldName, "_") // page_1234
		if !ok || kind != "page" || values[0] == "ignore" {
			continue
		}
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			views.RenderError(w, err)
			return
		}
		pageNum, err := srv.dreamer.DB().GetPageNum(r.Context(), id)
		if err != nil {
			views.RenderError(w, err)
			return
		}
		pagesToUse = append(pagesToUse, pageRef{
			ID:   id,
			Full: values[0] == "full",
			Num:  pageNum,
		})

	}

	// sort by num!
	slices.SortFunc(pagesToUse, func(a, b pageRef) int {
		return int(a.Num) - int(b.Num)
	})

	for _, page := range pagesToUse {
		if err := thread.AddPage(r.Context(), page.ID, page.Full); err != nil {
			views.RenderError(w, err)
			return
		}
	}

	if err := thread.User(r.Context(), message); err != nil {
		views.RenderError(w, err)
		return
	}

	err = srv.worker.Try(thread.Entity().ID, func(ctx context.Context) {
		_, err = thread.Run(ctx)
		if err != nil {
			slog.Error("failed to chat", "thread", threadID, "error", err)
		} else {
			if err := srv.dreamer.DB().UpdateChatDraft(ctx, dbo.UpdateChatDraftParams{
				Draft: "",
				ID:    thread.Entity().ID,
			}); err != nil {
				slog.Error("failed to update draft", "thread", threadID, "error", err)
			}
		}
	})
	if err != nil {
		views.RenderError(w, err)
		return
	}

	w.Header().Set("Location", ".#end")
	w.WriteHeader(http.StatusSeeOther)
}

func (srv *Chats) deleteMessage(w http.ResponseWriter, r *http.Request) {
	// threadID, _ := strconv.ParseInt(r.PathValue("threadID"), 10, 64)
	messageID, _ := strconv.ParseInt(r.PathValue("messageID"), 10, 64)

	if err := srv.dreamer.DB().DeleteMessage(r.Context(), messageID); err != nil {
		views.RenderError(w, err)
		return
	}
	if views.IsHTMX(r) {
		w.Header().Set("HX-Redirect", ".")
	} else {
		w.Header().Set("Location", "../../")
	}
	w.WriteHeader(http.StatusSeeOther)
}

func (srv *Chats) updateMessageContent(w http.ResponseWriter, r *http.Request) {
	messageID, _ := strconv.ParseInt(r.PathValue("messageID"), 10, 64)

	if err := srv.dreamer.DB().UpdateMessageContent(r.Context(), dbo.UpdateMessageContentParams{
		Content: r.FormValue("content"),
		ID:      messageID,
	}); err != nil {
		views.RenderError(w, err)
		return
	}
	if views.IsHTMX(r) {
		w.Header().Set("HX-Redirect", ".")
	} else {
		w.Header().Set("Location", "../../#m"+strconv.FormatInt(messageID, 10))
	}
	w.WriteHeader(http.StatusSeeOther)
}
