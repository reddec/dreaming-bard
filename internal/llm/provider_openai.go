package llm

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
	"github.com/openai/openai-go/packages/param"

	"github.com/reddec/dreaming-bard/internal/common"
)

func DefaultOpenAIProvider() OpenAIProvider {
	return OpenAIProvider{
		Model:       "gpt-4o",
		Timeout:     120 * time.Second,
		MaxTokens:   8192,
		Temperature: 1,
		TopP:        1,
	}
}

type OpenAIProvider struct {
	URL         string        `help:"OpenAI base URL" env:"URL" default:"https://api.openai.com/v1"`
	Model       string        `help:"OpenAI model name" env:"MODEL" default:"gpt-4o"`
	Token       string        `help:"OpenAI API token" env:"TOKEN"`
	Timeout     time.Duration `help:"Timeout" env:"TIMEOUT" default:"3m"`
	MaxTokens   int64         `help:"Max tokens" env:"MAX_TOKENS" default:"8192"`
	Temperature float64       `help:"Temperature" env:"TEMPERATURE" default:"0.8"`
	TopP        float64       `help:"Top P" env:"TOP_P" default:"0.9"`
}

func (op OpenAIProvider) Create() (ChatFunction, error) {
	cl := openai.NewClient(option.WithBaseURL(op.URL), option.WithAPIKey(op.Token), option.WithRequestTimeout(op.Timeout))
	return func(ctx context.Context, prompt Prompt) (*Result, error) {
		model := prompt.getModel(op.Model)
		logger := slog.With("provider", "openai", "model", model, "base_url", op.URL)

		var toos = make([]openai.ChatCompletionToolParam, 0, len(prompt.Tools))
		for _, f := range prompt.Tools {

			var description param.Opt[string]
			if f.description != "" {
				description = param.NewOpt(f.description)
			}
			toos = append(toos, openai.ChatCompletionToolParam{
				Function: openai.FunctionDefinitionParam{
					Name:        f.name,
					Description: description,
					Parameters:  f.inputSchema.ToOpenAPI(),
				},
			})
		}
		params := openai.ChatCompletionNewParams{
			Messages:            toOpenAI(prompt.System, prompt.History...),
			Model:               model,
			Tools:               toos,
			MaxCompletionTokens: openai.Int(op.MaxTokens),
			Temperature:         openai.Float(op.Temperature),
			TopP:                openai.Float(op.TopP),
		}
		var out string
		var iterateMore = true
		var stats Stats
		for iterateMore {
			iterateMore = false
			res, err := cl.Chat.Completions.New(ctx, params)
			if err != nil {
				return nil, fmt.Errorf("chat: %w", err)
			}
			stats.InputTokens += res.Usage.PromptTokens
			stats.OutputTokens += res.Usage.CompletionTokens
			prompt.safeNotifyStats(stats)

			out = res.Choices[0].Message.Content
			if out != "" {
				out = noThink(out)
				prompt.safeNotify(Message{Role: common.RoleAssistant, Content: out})
			}
			params.Messages = append(params.Messages, res.Choices[0].Message.ToParam())
			logger.Info("AI response", "content", out)

			for _, toolCall := range res.Choices[0].Message.ToolCalls {
				iterateMore = true
				logger.Info("tool call", "tool", toolCall.Function.Name, "args", toolCall.Function.Arguments, "id", toolCall.ID)
				prompt.safeNotify(Message{Role: common.RoleToolCall, Content: toolCall.Function.Arguments, Tool: toolCall.Function.Name, ToolID: toolCall.ID})
				tool := prompt.toolByName(toolCall.Function.Name)

				if tool == nil {
					logger.Error("failed to find tool", "function", toolCall.Function.Name)
					prompt.safeNotify(Message{Role: common.RoleToolResult, Content: "tool not found", Tool: toolCall.Function.Name, ToolID: toolCall.ID})
					params.Messages = append(params.Messages, openai.ToolMessage("function "+toolCall.Function.Name+" not found", toolCall.ID))
					continue
				}

				toolRes, err := tool.CallJSON(ctx, json.RawMessage(toolCall.Function.Arguments))
				if err != nil {
					logger.Error("failed to call tool", "function", toolCall.Function.Name, "error", err)
					prompt.safeNotify(Message{Role: common.RoleToolResult, Content: err.Error(), Tool: toolCall.Function.Name, ToolID: toolCall.ID})
					params.Messages = append(params.Messages, openai.ToolMessage("function call failed: "+err.Error(), toolCall.ID))
					continue
				}

				prompt.safeNotify(Message{Role: common.RoleToolResult, Content: string(toolRes), Tool: toolCall.Function.Name, ToolID: toolCall.ID})
				params.Messages = append(params.Messages, openai.ToolMessage(string(toolRes), toolCall.ID))
			}
		}
		return &Result{
			Content: out,
			Stats:   stats,
		}, nil
	}, nil
}
func toOpenAI(system string, messages ...Message) []openai.ChatCompletionMessageParamUnion {
	ans := make([]openai.ChatCompletionMessageParamUnion, 0, 1+len(messages))
	if system != "" {
		ans = append(ans, openai.ChatCompletionMessageParamUnion{
			OfSystem: &openai.ChatCompletionSystemMessageParam{
				Content: openai.ChatCompletionSystemMessageParamContentUnion{
					OfString: param.NewOpt(system),
				},
			},
		})
	}
	for _, m := range messages {
		ans = append(ans, messageToOpenAI(m))
	}
	return ans
}

func messageToOpenAI(m Message) openai.ChatCompletionMessageParamUnion {
	switch m.Role {
	case common.RoleAssistant:
		return openai.ChatCompletionMessageParamUnion{
			OfAssistant: &openai.ChatCompletionAssistantMessageParam{
				Content: openai.ChatCompletionAssistantMessageParamContentUnion{
					OfString: param.NewOpt(m.Content),
				},
			},
		}
	case common.RoleToolCall:
		return openai.ChatCompletionMessageParamUnion{
			OfAssistant: &openai.ChatCompletionAssistantMessageParam{
				Content: openai.ChatCompletionAssistantMessageParamContentUnion{
					OfString: param.NewOpt(m.Content),
				},
				ToolCalls: []openai.ChatCompletionMessageToolCallParam{
					{ID: m.ToolID, Function: openai.ChatCompletionMessageToolCallFunctionParam{
						Arguments: m.Content,
						Name:      m.Tool,
					}},
				},
			},
		}
	case common.RoleToolResult:
		return openai.ChatCompletionMessageParamUnion{
			OfTool: &openai.ChatCompletionToolMessageParam{
				Content: openai.ChatCompletionToolMessageParamContentUnion{
					OfString: param.NewOpt(m.Content),
				},
				ToolCallID: m.ToolID,
			},
		}
	case common.RoleUser:
		fallthrough
	default:
		return openai.ChatCompletionMessageParamUnion{
			OfUser: &openai.ChatCompletionUserMessageParam{
				Content: openai.ChatCompletionUserMessageParamContentUnion{
					OfString: param.NewOpt(m.Content),
				},
			},
		}
	}
}
