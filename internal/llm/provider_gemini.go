package llm

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"google.golang.org/genai"

	"github.com/reddec/dreaming-bard/internal/common"
)

func DefaultGeminiProvider() GeminiProvider {
	return GeminiProvider{
		Model:       "gemini-2.5-flash",
		Timeout:     120 * time.Second,
		MaxTokens:   8192,
		Temperature: 0.8,
		TopP:        0.9,
		TopK:        40,
		ThresholdSettings: ThresholdSettings{
			Harassment:       "NONE",
			HateSpeech:       "NONE",
			SexuallyExplicit: "NONE",
			DangerousContent: "NONE",
		},
	}
}

type ThresholdSettings struct {
	Harassment       string `help:"Harassment threshold" env:"HARASSMENT" default:"NONE" yaml:"harassment"`
	HateSpeech       string `help:"Hate speech threshold" env:"HATE_SPEECH" default:"NONE" yaml:"hate_speech"`
	SexuallyExplicit string `help:"Explicit content" env:"EXPLICIT" default:"NONE" yaml:"sexually_explicit"`
	DangerousContent string `help:"Dangerous content threshold" env:"DANGEROUS_CONTENT" default:"NONE" yaml:"dangerous_content"`
}

type GeminiProvider struct {
	Model             string            `help:"Gemini model name" env:"MODEL" default:"gemini-2.0-flash"`
	Token             string            `help:"Google AI API key" env:"TOKEN"`
	Timeout           time.Duration     `help:"Timeout" env:"TIMEOUT" default:"120s"`
	MaxTokens         int64             `help:"Max tokens" env:"MAX_TOKENS" default:"8192"`
	Temperature       float64           `help:"Temperature" env:"TEMPERATURE" default:"0.8"`
	TopP              float64           `help:"Top P" env:"TOP_P" default:"0.9"`
	TopK              int64             `help:"Top K" env:"TOP_K" default:"40"`
	ThresholdSettings ThresholdSettings `embed:"" prefix:"threshold-" envprefix:"THRESHOLD_"`
}

func convertHarmBlockThreshold(thresholdLevel string) genai.HarmBlockThreshold {
	switch thresholdLevel {
	case "NONE":
		return genai.HarmBlockThresholdBlockNone
	case "LOW":
		return genai.HarmBlockThresholdBlockLowAndAbove
	case "MEDIUM":
		return genai.HarmBlockThresholdBlockMediumAndAbove
	case "HIGH":
		return genai.HarmBlockThresholdBlockOnlyHigh
	case "UNSPECIFIED":
		fallthrough
	default:
		return genai.HarmBlockThresholdUnspecified
	}
}

func (gp GeminiProvider) configureSafetySettings() []*genai.SafetySetting {
	return []*genai.SafetySetting{
		{Category: genai.HarmCategoryHarassment, Threshold: convertHarmBlockThreshold(gp.ThresholdSettings.Harassment)},
		{Category: genai.HarmCategoryHateSpeech, Threshold: convertHarmBlockThreshold(gp.ThresholdSettings.HateSpeech)},
		{Category: genai.HarmCategorySexuallyExplicit, Threshold: convertHarmBlockThreshold(gp.ThresholdSettings.SexuallyExplicit)},
		{Category: genai.HarmCategoryDangerousContent, Threshold: convertHarmBlockThreshold(gp.ThresholdSettings.DangerousContent)},
	}
}

func (gp GeminiProvider) Create() (ChatFunction, error) {
	ctx := context.Background()
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey:  gp.Token,
		Backend: genai.BackendGeminiAPI,
	})
	if err != nil {
		return nil, fmt.Errorf("create gemini client: %w", err)
	}

	return func(ctx context.Context, prompt Prompt) (*Result, error) {
		model := prompt.getModel(gp.Model)
		logger := slog.With("provider", "gemini", "model", model)

		var config genai.GenerateContentConfig
		config.MaxOutputTokens = int32(gp.MaxTokens)
		config.Temperature = ref(float32(gp.Temperature))
		config.TopP = ref(float32(gp.TopP))
		config.TopK = ref(float32(gp.TopK))
		config.SafetySettings = gp.configureSafetySettings()

		if prompt.System != "" {
			config.SystemInstruction = &genai.Content{
				Parts: []*genai.Part{{Text: prompt.System}},
			}
		}

		if len(prompt.Tools) > 0 {
			var functions = make([]*genai.FunctionDeclaration, 0, len(prompt.Tools))
			for _, tool := range prompt.Tools {
				functions = append(functions, &genai.FunctionDeclaration{
					Name:        tool.name,
					Description: tool.description,
					Parameters:  toGeminiSchema(tool.inputSchema),
				})
			}
			config.Tools = []*genai.Tool{
				{
					FunctionDeclarations: functions,
				},
			}
		}
		logger.Info("available tools", "tools", len(prompt.Tools))
		var history = make([]*genai.Content, 0, len(prompt.History))

		for _, msg := range prompt.History {
			history = append(history, messageToGemini(msg))
		}

		var finalResponse string
		resp, err := client.Models.GenerateContent(ctx, model, history, &config)
		if err != nil {
			return nil, fmt.Errorf("send message: %w", err)
		}

		var stats Stats

		for {
			stats.InputTokens += int64(resp.UsageMetadata.PromptTokenCount)
			stats.OutputTokens += int64(resp.UsageMetadata.CandidatesTokenCount)
			prompt.safeNotifyStats(stats)
			if len(resp.Candidates) == 0 {
				return nil, fmt.Errorf("no response candidates")
			}
			candidate := resp.Candidates[0]
			history = append(history, candidate.Content)

			for _, part := range candidate.Content.Parts {
				if txt := part.Text; txt != "" {
					finalResponse = txt
					logger.Info("AI response", "content", finalResponse)
				}
			}
			if finalResponse != "" {
				prompt.safeNotify(Message{Role: common.RoleAssistant, Content: finalResponse})
			}
			var responseParts []*genai.Part
			for _, part := range candidate.Content.Parts {

				if funcCall := part.FunctionCall; funcCall != nil {
					logger.Info("tool call", "tool", funcCall.Name, "args", funcCall.Args)
					argsJSON, _ := json.Marshal(funcCall.Args)

					prompt.safeNotify(Message{Role: common.RoleToolCall, Content: string(argsJSON), Tool: funcCall.Name, ToolID: funcCall.ID})
					tool := prompt.toolByName(funcCall.Name)

					if tool == nil {
						logger.Error("failed to find tool", "function", funcCall.Name)
						prompt.safeNotify(Message{Role: common.RoleToolResult, Content: `{"result":"tool not found"}`, Tool: funcCall.Name, ToolID: funcCall.ID})

						responseParts = append(responseParts, &genai.Part{FunctionResponse: &genai.FunctionResponse{
							ID:   funcCall.ID,
							Name: funcCall.Name,
							Response: map[string]any{
								"error": fmt.Sprintf("function %s not found", funcCall.Name),
							},
						}})
						continue
					}

					if err != nil {
						logger.Error("failed to marshal args", "function", funcCall.Name, "error", err)
						response := map[string]any{
							"error": fmt.Sprintf("failed to marshal arguments: %v", err),
						}
						prompt.safeNotify(Message{Role: common.RoleToolResult, Content: string(mustJsonResponse(response)), Tool: funcCall.Name, ToolID: funcCall.ID})
						responseParts = append(responseParts, &genai.Part{FunctionResponse: &genai.FunctionResponse{
							ID:       funcCall.ID,
							Name:     funcCall.Name,
							Response: response,
						}})
						continue
					}

					toolResult, err := tool.CallJSON(ctx, argsJSON)
					if err != nil {
						logger.Error("failed to call tool", "function", funcCall.Name, "error", err)
						response := map[string]any{
							"error": fmt.Sprintf("function call failed: %v", err),
						}
						prompt.safeNotify(Message{Role: common.RoleToolResult, Content: string(mustJsonResponse(response)), Tool: funcCall.Name, ToolID: funcCall.ID})
						responseParts = append(responseParts, &genai.Part{FunctionResponse: &genai.FunctionResponse{
							ID:       funcCall.ID,
							Name:     funcCall.Name,
							Response: response,
						}})
						continue
					}

					var result map[string]any
					if err := json.Unmarshal(toolResult, &result); err != nil {
						// If not JSON, treat as string
						result = map[string]any{"result": toolResult}
					}
					prompt.safeNotify(Message{Role: common.RoleToolResult, Content: string(mustJsonResponse(result)), Tool: funcCall.Name, ToolID: funcCall.ID})

					responseParts = append(responseParts, &genai.Part{FunctionResponse: &genai.FunctionResponse{
						ID:       funcCall.ID,
						Name:     funcCall.Name,
						Response: result,
					}})
				}
			}
			if len(responseParts) > 0 {
				history = append(history, &genai.Content{
					Parts: responseParts,
				})
				resp, err = client.Models.GenerateContent(ctx, model, history, &config)
			} else {
				break
			}
		}

		return &Result{
			Content: finalResponse,
			Stats:   stats,
		}, nil
	}, nil
}

func toGeminiSchema(schema *Schema) *genai.Schema {
	if schema == nil {
		return &genai.Schema{Type: genai.TypeObject}
	}
	geminiSchema := &genai.Schema{
		Description: schema.Description,
	}

	switch schema.Type {
	case TypeObject:
		geminiSchema.Type = genai.TypeObject
		if len(schema.Fields) == 0 {
			break
		}
		geminiSchema.Properties = make(map[string]*genai.Schema)
		geminiSchema.Required = make([]string, 0, len(schema.Fields))

		for fieldName, fieldSchema := range schema.Fields {
			geminiSchema.Properties[fieldName] = toGeminiSchema(fieldSchema)
			geminiSchema.Required = append(geminiSchema.Required, fieldName)
		}
	case TypeArray:
		geminiSchema.Type = genai.TypeArray
		if schema.Items != nil {
			geminiSchema.Items = toGeminiSchema(schema.Items)
		}
	case TypeBoolean:
		geminiSchema.Type = genai.TypeBoolean
	case TypeInteger:
		geminiSchema.Type = genai.TypeInteger
	case TypeNumber:
		geminiSchema.Type = genai.TypeNumber
	case TypeString:
		geminiSchema.Type = genai.TypeString
	default:
		// fallback
		geminiSchema.Type = genai.TypeObject
	}

	return geminiSchema
}

func messageToGemini(msg Message) *genai.Content {
	switch msg.Role {
	case common.RoleAssistant:
		return &genai.Content{
			Role:  "model",
			Parts: []*genai.Part{{Text: msg.Content}},
		}
	case common.RoleToolCall:
		var args map[string]any
		if err := json.Unmarshal([]byte(msg.Content), &args); err != nil {
			args = map[string]any{"data": msg.Content} // hack
		}
		return &genai.Content{
			Role: "model",
			Parts: []*genai.Part{{
				FunctionCall: &genai.FunctionCall{
					ID:   msg.ToolID,
					Name: msg.Tool,
					Args: args,
				}},
			},
		}
	case common.RoleToolResult:
		var result map[string]any
		if err := json.Unmarshal([]byte(msg.Content), &result); err != nil {
			result = map[string]any{"result": msg.Content}
		}
		return &genai.Content{
			Role: "user",
			Parts: []*genai.Part{{
				FunctionResponse: &genai.FunctionResponse{
					ID:       msg.ToolID,
					Name:     msg.Tool,
					Response: result,
				}},
			},
		}
	case common.RoleUser:
		fallthrough
	default:
		return &genai.Content{
			Role:  "user",
			Parts: []*genai.Part{{Text: msg.Content}},
		}
	}
}

func mustJsonResponse(result any) json.RawMessage {
	v, err := json.Marshal(result)
	if err != nil {
		panic(err)
	}
	return v
}

func ref[T any](v T) *T {
	return &v
}
