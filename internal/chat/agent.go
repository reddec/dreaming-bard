package chat

import (
	"context"
	"fmt"

	"github.com/reddec/dreaming-bard/internal/llm"
	"github.com/reddec/dreaming-bard/internal/utils/events"
)

func NewAgent(system string, model llm.ChatFunction) *Agent {
	return &Agent{
		system: system,
		model:  model,
	}
}

type Agent struct {
	history       []llm.Message
	system        string
	modelName     string
	tools         []*llm.Function
	model         llm.ChatFunction
	stats         llm.Stats
	onMessage     events.Emitter[llm.Message]
	onStats       events.Emitter[llm.Stats]
	onUserMessage events.Emitter[llm.Message]
}

func (a *Agent) OnStats() events.Event[llm.Stats] {
	return &a.onStats
}

func (a *Agent) OnMessage() events.Event[llm.Message] {
	return &a.onMessage
}

func (a *Agent) OnUserMessage() events.Event[llm.Message] {
	return &a.onUserMessage
}

func (a *Agent) AddTool(tool *llm.Function) *Agent {
	a.tools = append(a.tools, tool)
	return a
}

func (a *Agent) Add(history ...llm.Message) *Agent {
	a.history = append(a.history, history...)
	return a
}

func (a *Agent) AddStats(stats ...llm.Stats) *Agent {
	for _, s := range stats {
		a.stats.Add(s)
	}
	return a
}

func (a *Agent) User(message string) *Agent {
	usr := llm.User(message)
	a.onUserMessage.Emit(usr)
	a.history = append(a.history, usr)
	return a
}

func (a *Agent) Model(model string) *Agent {
	a.modelName = model
	return a
}

func (a *Agent) Assistant(message string) *Agent {
	a.history = append(a.history, llm.Assistant(message))
	return a
}

func (a *Agent) History() []llm.Message {
	return a.history
}

func (a *Agent) Run(ctx context.Context) (string, error) {
	res, err := a.model(ctx, llm.Prompt{
		System:  a.system,
		History: a.history,
		Tools:   a.tools,
		Model:   a.modelName,
		Update: func(msg llm.Message) {
			a.onMessage.Emit(msg)
			a.history = append(a.history, msg)
		},
		Stats: func(stats llm.Stats) {
			a.onStats.Emit(stats)
			a.stats.Add(stats)
		},
	})

	if err != nil {
		return "", fmt.Errorf("chat: %w", err)
	}

	return res.Content, nil
}
