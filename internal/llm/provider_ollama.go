package llm

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/ollama/ollama/api"

	"github.com/reddec/dreaming-bard/internal/common"
)

func Local(model string) OllamaProvider {
	return OllamaProvider{
		URL:         "http://localhost:11434",
		Model:       model,
		Timeout:     120 * time.Second,
		ContextSize: 32768,
		MaxTokens:   32768,
		Temperature: 0.95,
		TopP:        20,
		TopK:        20,
		MinP:        0,
		NoThink:     false,
	}
}

func DefaultOllama() OllamaProvider {
	return Local("qwen3:14b")
}

type OllamaProvider struct {
	URL         string        `help:"Ollama OpenAPI URL" env:"URL" default:"http://localhost:11434" yaml:"url"`
	Model       string        `help:"Ollama model name" env:"MODEL" default:"qwen3:14b" yaml:"model"`
	Timeout     time.Duration `help:"Timeout" env:"TIMEOUT" default:"120s" yaml:"timeout"`
	ContextSize int64         `help:"Context size" env:"CONTEXT_SIZE" default:"32768" yaml:"context_size"`
	MaxTokens   int64         `help:"Max tokens" env:"MAX_TOKENS" default:"32768" yaml:"max_tokens"`
	Temperature float64       `help:"Temperature" env:"TEMPERATURE" default:"0.6" yaml:"temperature"`
	TopP        float64       `help:"Top P" env:"TOP_P" default:"0.95" yaml:"top_p"`
	TopK        int64         `help:"Top K" env:"TOP_K" default:"20" yaml:"top_k"`
	MinP        float64       `help:"Min P" env:"MIN_P" default:"0" yaml:"min_p"`
	NoThink     bool          `help:"Disable thinking" env:"NO_THINK" yaml:"no_think"`
}

func (op OllamaProvider) Create() (ChatFunction, error) {
	u, err := url.Parse(op.URL)
	if err != nil {
		return nil, fmt.Errorf("parse url: %w", err)
	}

	cl := api.NewClient(u, &http.Client{
		Timeout: op.Timeout,
	})
	think := !op.NoThink
	return func(ctx context.Context, prompt Prompt) (*Result, error) {
		model := prompt.getModel(op.Model)
		logger := slog.With("provider", "ollama", "model", model)

		var mapped []api.Message
		if prompt.System != "" {
			mapped = append(mapped, api.Message{
				Role:    "system",
				Content: prompt.System,
			})
		}
		for _, m := range prompt.History {
			mapped = append(mapped, api.Message{
				Role:    string(m.Role),
				Content: m.Content,
			})
			switch m.Role {
			case common.RoleUser:
				mapped = append(mapped, api.Message{
					Role:    "user",
					Content: m.Content,
				})
			case common.RoleAssistant:
				mapped = append(mapped, api.Message{
					Role:    "assistant",
					Content: m.Content,
				})
			case common.RoleToolCall:
				var args map[string]any
				if err := json.Unmarshal([]byte(m.Content), &args); err != nil {
					return nil, fmt.Errorf("decode tool call arguments: %w", err)
				}
				var tf = api.ToolCallFunction{
					Index:     0,
					Name:      m.Tool,
					Arguments: args,
				}
				mapped = append(mapped, api.Message{
					Role:      "assistant",
					ToolCalls: []api.ToolCall{{Function: tf}},
				})
			case common.RoleToolResult:
				mapped = append(mapped, api.Message{
					Role:    "tool",
					Content: m.Content,
				})
			}
		}

		var llmTools []api.Tool
		for _, t := range prompt.Tools {
			llmTools = append(llmTools, api.Tool{
				Type:     "function",
				Function: mapFuncToOllama(t),
			})
		}
		var out string
		var iterateMore = true
		var stats Stats

		for iterateMore {
			iterateMore = false

			var res api.ChatResponse
			err = cl.Chat(ctx, &api.ChatRequest{
				Model:    model,
				Messages: mapped,
				Stream:   new(bool),
				Tools:    llmTools,
				Options: map[string]any{
					"num_ctx":     op.ContextSize,
					"num_predict": op.MaxTokens,
					"temperature": op.Temperature,
					"top_k":       op.TopK,
					"top_p":       op.TopP,
					"min_p":       op.MinP,
				},
				Think: &think,
			}, func(response api.ChatResponse) error {
				res = response
				return nil
			})

			if err != nil {
				return nil, fmt.Errorf("chat: %w", err)
			}
			stats.InputTokens += int64(res.PromptEvalCount)
			stats.OutputTokens += int64(res.EvalCount)
			prompt.safeNotifyStats(stats)

			out = res.Message.Content
			mapped = append(mapped, res.Message)

			if out != "" {
				prompt.safeNotify(Message{Role: common.RoleAssistant, Content: noThink(out)})
			}

			for _, toolCall := range res.Message.ToolCalls {
				iterateMore = true
				toolID := strconv.Itoa(toolCall.Function.Index)

				// hack to keep interface and track
				raw, err := json.Marshal(toolCall.Function.Arguments)
				if err != nil {
					return nil, fmt.Errorf("marshal tool call arguments: %w", err)
				}

				prompt.safeNotify(Message{Role: common.RoleToolCall, Content: string(raw), Tool: toolCall.Function.Name, ToolID: toolID})

				tool := prompt.toolByName(toolCall.Function.Name)
				if tool == nil {
					prompt.safeNotify(Message{Role: common.RoleToolResult, Content: "tool not found", Tool: toolCall.Function.Name, ToolID: toolID})
					logger.Error("failed to find tool", "function", toolCall.Function.Name)
					mapped = append(mapped, api.Message{
						Role:    "tool",
						Content: "function " + toolCall.Function.Name + " not found",
					})
					continue
				}

				toolRes, err := tool.CallJSON(ctx, raw)
				if err != nil {
					logger.Error("failed to call tool", "function", toolCall.Function.Name, "error", err)
					prompt.safeNotify(Message{Role: common.RoleToolResult, Content: err.Error(), Tool: toolCall.Function.Name, ToolID: toolID})

					mapped = append(mapped, api.Message{
						Role:    "tool",
						Content: "function" + toolCall.Function.Name + " call failed: " + err.Error(),
					})
					continue
				}
				prompt.safeNotify(Message{Role: common.RoleToolResult, Content: string(toolRes), Tool: toolCall.Function.Name, ToolID: toolID})

				mapped = append(mapped, api.Message{
					Role:    "tool",
					Content: string(toolRes),
				})
			}
		}

		return &Result{
			Content: noThink(out),
			Stats:   stats,
		}, nil
	}, nil
}

func noThink(value string) string {
	a, b, _ := strings.Cut(value, "</think>")
	if b != "" {
		return strings.TrimSpace(b)
	}
	return a
}

func mapFuncToOllama(f *Function) api.ToolFunction {
	var tool api.ToolFunction
	tool.Name = f.name
	tool.Description = f.description
	tool.Parameters.Type = "object"
	tool.Parameters.Properties = make(map[string]struct {
		Type        api.PropertyType `json:"type"`
		Items       any              `json:"items,omitempty"`
		Description string           `json:"description"`
		Enum        []any            `json:"enum,omitempty"`
	}, len(f.inputSchema.Fields))
	tool.Parameters.Required = make([]string, 0, len(f.inputSchema.Fields))
	for field, schema := range f.inputSchema.Fields {
		x := tool.Parameters.Properties[field]
		x.Description = schema.Description
		switch schema.Type {
		case TypeString:
			x.Type = []string{"string"}
		case TypeArray:
			x.Type = []string{"array"}
			x.Items = schema.ToOpenAPI()
		case TypeObject:
			x.Type = []string{"object"}
		case TypeBoolean:
			x.Type = []string{"boolean"}
		case TypeInteger:
			x.Type = []string{"integer"}
		case TypeNumber:
			x.Type = []string{"number"}
		}
		tool.Parameters.Properties[field] = x
		tool.Parameters.Required = append(tool.Parameters.Required, field)
	}

	return tool
}
