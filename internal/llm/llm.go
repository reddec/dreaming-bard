package llm

import (
	"context"

	"github.com/reddec/dreaming-bard/internal/common"
)

type Message struct {
	Role    common.Role
	Content string `yaml:"content,omitempty" json:"content,omitempty"`
	Tool    string `yaml:"tool,omitempty" json:"tool,omitempty"`
	ToolID  string `yaml:"tool_id,omitempty" json:"tool_id,omitempty"`
}

type Stats struct {
	InputTokens  int64 `yaml:"input_tokens" json:"input_tokens"`
	OutputTokens int64 `yaml:"output_tokens" json:"output_tokens"`
}

func (s *Stats) Add(other Stats) {
	s.InputTokens += other.InputTokens
	s.OutputTokens += other.OutputTokens
}

type Prompt struct {
	System  string
	Model   string // override model
	History []Message
	Tools   []*Function
	Update  func(msg Message)
	Stats   func(stats Stats)
}

func (p *Prompt) safeNotify(msg Message) {
	if p.Update != nil {
		p.Update(msg)
	}
}

func (p *Prompt) safeNotifyStats(stats Stats) {
	if p.Stats != nil {
		p.Stats(stats)
	}
}

func (p *Prompt) toolByName(name string) *Function {
	for _, t := range p.Tools {
		if t.name == name {
			return t
		}
	}
	return nil
}

func (p *Prompt) getModel(defaultModel string) string {
	if p.Model == "" {
		return defaultModel
	}
	return p.Model
}

type Result struct {
	Content string
	Stats   Stats
}

type ChatFunction func(ctx context.Context, prompt Prompt) (*Result, error)

func User(value string) Message {
	return Message{
		Role:    common.RoleUser,
		Content: value,
	}
}

func Assistant(value string) Message {
	return Message{
		Role:    common.RoleAssistant,
		Content: value,
	}
}
