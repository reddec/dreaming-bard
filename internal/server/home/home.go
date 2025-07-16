package home

import (
	"context"
	"net/http"

	"github.com/reddec/dreaming-bard/internal/dreamwriter"
	"github.com/reddec/dreaming-bard/internal/server/views"
)

func New(dreamer *dreamwriter.DreamWriter) *Home {
	mux := http.NewServeMux()
	srv := &Home{
		Handler:   mux,
		dreamer:   dreamer,
		lastPages: 5,
		lastChats: 3,
	}
	mux.HandleFunc("GET /", srv.index)

	return srv
}

type Home struct {
	http.Handler
	lastPages int64
	lastChats int64
	dreamer   *dreamwriter.DreamWriter
}

func (srv *Home) Run(ctx context.Context) error {
	// TODO: background tasks
	return nil
}

func (srv *Home) index(w http.ResponseWriter, r *http.Request) {
	prompts, err := srv.dreamer.DB().ListPinnedPrompts(r.Context())
	if err != nil {
		views.RenderError(w, err)
		return
	}
	lastPages, err := srv.dreamer.DB().ListLastPages(r.Context(), srv.lastPages)
	if err != nil {
		views.RenderError(w, err)
		return
	}
	lastChats, err := srv.dreamer.DB().ListLastChats(r.Context(), srv.lastChats)
	if err != nil {
		views.RenderError(w, err)
		return
	}
	viewIndex().HTML(w, indexParams{
		Config:        srv.dreamer.Provider(),
		PinnedPrompts: prompts,
		Pages:         lastPages,
		Chats:         lastChats,
	})
}
