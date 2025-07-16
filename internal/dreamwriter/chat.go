package dreamwriter

import (
	"context"
	"encoding/xml"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/reddec/dreaming-bard/internal/dbo"
	"github.com/reddec/dreaming-bard/internal/llm"
)

const (
	MaxContent = 255
)

func makeChat(db *dbo.Queries, provider llm.Provider, chat dbo.Chat) *Chat {
	return &Chat{
		db:       db,
		provider: provider,
		chat:     chat,
	}
}

type document struct {
	xml.Name `xml:"document"`
	Title    string `xml:"title,attr"`
	Num      int64  `xml:"num,attr,omitempty"`
	Category string `xml:"category,attr,omitempty"`
	Content  string `xml:",chardata"`
}

type Chat struct {
	db       *dbo.Queries
	provider llm.Provider
	chat     dbo.Chat

	contextBlocks []document
}

func (c *Chat) User(ctx context.Context, message string) error {
	if len(c.contextBlocks) > 0 {
		contextData, err := xml.MarshalIndent(c.contextBlocks, "", "  ")
		if err != nil {
			return fmt.Errorf("marshal context blocks: %w", err)
		}
		c.contextBlocks = nil
		message += "\n\n<context>\n" + string(contextData) + "\n</context>"
	}

	return c.AddMessages(ctx, llm.User(message))
}

func (c *Chat) AddMessages(ctx context.Context, messages ...llm.Message) error {
	if len(messages) == 0 {
		return nil
	}
	return c.db.Transaction(ctx, func(q *dbo.Queries) error {
		for _, msg := range messages {
			_, err := q.CreateMessage(ctx, dbo.CreateMessageParams{
				ChatID:   c.chat.ID,
				Content:  msg.Content,
				Role:     msg.Role,
				ToolID:   msg.ToolID,
				ToolName: msg.Tool,
			})
			if err != nil {
				return err
			}
		}
		return nil
	})
}

func (c *Chat) AddContexts(ctx context.Context, ids ...int64) error {
	for _, ctID := range ids {
		doc, err := c.db.GetContext(ctx, ctID)
		if err != nil {
			return fmt.Errorf("get context: %w", err)
		}
		c.contextBlocks = append(c.contextBlocks, document{
			Title:    doc.Title,
			Category: doc.Category,
			Content:  doc.Content,
		})
	}
	return nil
}

func (c *Chat) AddDocument(_ context.Context, description string, content string) error {
	c.contextBlocks = append(c.contextBlocks, document{
		Category: strings.TrimSpace(description),
		Content:  content,
	})
	return nil
}

func (c *Chat) AddPage(ctx context.Context, id int64, inline bool) error {
	page, err := c.db.GetPage(ctx, id)
	if err != nil {
		return fmt.Errorf("get page: %w", err)
	}
	content := page.Content
	kind := "full content"

	if !inline {
		if page.Summary != "" {
			kind = "summary"
			content = page.Summary
		} else if len([]rune(content)) <= MaxContent {
			kind = "full content"
		} else {
			content = string([]rune(content)[:MaxContent]) + "..."
			kind = "partial (truncated) content"
		}
	}
	c.contextBlocks = append(c.contextBlocks, document{
		Title:    fmt.Sprintf("%s of page #%d", kind, page.Num),
		Category: "page",
		Content:  content,
	})
	return nil
}

func (c *Chat) AddNote(_ context.Context, note string) error {
	c.contextBlocks = append(c.contextBlocks, document{
		Category: "note",
		Content:  note,
	})
	return nil
}

func (c *Chat) Run(ctx context.Context, options ...RunOption) (string, error) {
	opts := runOptions{}
	for _, opt := range options {
		opt(&opts)
	}

	role, err := c.db.GetRole(ctx, c.chat.RoleID)
	if err != nil {
		return "", fmt.Errorf("load role: %w", err)
	}

	history, err := c.getHistory(ctx)
	if err != nil {
		return "", fmt.Errorf("load history: %w", err)
	}

	if opts.prefill != "" {
		history = append(history, llm.Assistant(opts.prefill))
	}

	provider, err := c.provider.Create()
	if err != nil {
		return "", fmt.Errorf("create provider: %w", err)
	}

	started := time.Now()
	slog.Info("chat started", "chat_id", c.chat.ID)
	res, err := provider(ctx, llm.Prompt{
		System:  strings.TrimSpace(role.System),
		History: history,
		Model:   role.Model,
		Update: func(msg llm.Message) {
			if err := c.AddMessages(ctx, msg); err != nil {
				slog.Error("failed to save message", "error", err)
			}
		},
		Stats: func(stats llm.Stats) {
			if err := c.db.AddChatStats(ctx, dbo.AddChatStatsParams{
				InputTokens:  stats.InputTokens,
				OutputTokens: stats.OutputTokens,
				ID:           c.chat.ID,
			}); err != nil {
				slog.Error("failed to save stats", "error", err)
			}
		},
	})

	if err != nil {
		return "", fmt.Errorf("chat: %w", err)
	}
	slog.Info("run complete",
		"chat_id", c.chat.ID,
		"role_id", role.ID,
		"role", role.Name,
		"role_model", role.Model,
		"num_messages", len(history),
		"input_tokens", res.Stats.InputTokens,
		"output_tokens", res.Stats.OutputTokens,
		"duration", time.Since(started),
		"annotation", c.chat.Annotation)
	return res.Content, nil
}

func (c *Chat) Entity() dbo.Chat {
	return c.chat
}

func (c *Chat) getHistory(ctx context.Context) ([]llm.Message, error) {
	messages, err := c.db.ListMessagesByChat(ctx, c.chat.ID)
	if err != nil {
		return nil, fmt.Errorf("load messages: %w", err)
	}

	var history = make([]llm.Message, 0, len(messages))
	for _, m := range messages {
		history = append(history, llm.Message{
			Role:    m.Role,
			Content: m.Content,
			Tool:    m.ToolName,
			ToolID:  m.ToolID,
		})
	}
	return history, nil
}

type runOptions struct {
	prefill string
}
type RunOption func(*runOptions)

func Prefill(prefill string) RunOption {
	return func(opts *runOptions) {
		opts.prefill = prefill
	}
}
