package dreamwriter

import (
	"context"
	"fmt"

	"github.com/reddec/dreaming-bard/internal/dbo"
	"github.com/reddec/dreaming-bard/internal/llm"
)

// Kinda dump of use-cases, ungrouped for now.

func NewDreamWriter(db *dbo.Queries, config llm.Provider) *DreamWriter {
	return &DreamWriter{
		db:       db,
		provider: config,
	}
}

type DreamWriter struct {
	db       *dbo.Queries
	provider llm.Provider
}

func (dw *DreamWriter) Provider() llm.Provider {
	return dw.provider
}

// Summarise generates a concise summary of the provided content using the specific LLM role (agent).
func (dw *DreamWriter) Summarise(ctx context.Context, roleID int64, content string) (string, error) {
	chat, err := dw.NewChat(ctx, roleID, Annotation("generate summary"))
	if err != nil {
		return "", fmt.Errorf("create chat: %w", err)
	}
	if err := chat.User(ctx, content); err != nil {
		return "", fmt.Errorf("add message: %w", err)
	}
	return chat.Run(ctx)
}

func (dw *DreamWriter) DB() *dbo.Queries {
	return dw.db
}

// NewChat - creates a new chat. I know, shocking revelation.
// This function does exactly what you might expect: creates new records in storage about chat
func (dw *DreamWriter) NewChat(ctx context.Context, roleID int64, options ...ChatOption) (*Chat, error) {
	params := newChatOpts{params: dbo.CreateChatParams{
		RoleID: roleID,
	}}
	for _, opt := range options {
		opt(&params)
	}

	dbChat, err := dw.db.CreateChat(ctx, params.params)
	if err != nil {
		return nil, fmt.Errorf("create chat: %w", err)
	}
	return makeChat(dw.db, dw.provider, dbChat), nil
}

func (dw *DreamWriter) OpenChat(ctx context.Context, chatID int64) (*Chat, error) {
	dbChat, err := dw.db.GetChat(ctx, chatID)
	if err != nil {
		return nil, fmt.Errorf("load chat: %w", err)
	}
	return makeChat(dw.db, dw.provider, dbChat), nil
}

type newChatOpts struct {
	params dbo.CreateChatParams
}

type ChatOption func(*newChatOpts)

func Annotation(annotation string) ChatOption {
	return func(opts *newChatOpts) {
		opts.params.Annotation = annotation
	}
}

func Draft(content string) ChatOption {
	return func(opts *newChatOpts) {
		opts.params.Draft = content
	}
}
