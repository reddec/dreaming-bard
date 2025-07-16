package llm

import (
	"fmt"
)

type Provider struct {
	Type   string         `help:"Provider type" env:"TYPE" default:"ollama" enum:"ollama,openai,gemini"`
	OpenAI OpenAIProvider `embed:"" prefix:"openai-" envprefix:"OPENAI_"`
	Ollama OllamaProvider `embed:"" prefix:"ollama-" envprefix:"OLLAMA_"`
	Gemini GeminiProvider `embed:"" prefix:"gemini-" envprefix:"GEMINI_"`
}

func (p *Provider) Create() (ChatFunction, error) {
	switch p.Type {
	case "ollama":
		return p.Ollama.Create()
	case "openai":
		return p.OpenAI.Create()
	case "gemini":
		return p.Gemini.Create()
	default:
		return nil, fmt.Errorf("unknown provider type: %s", p.Type)
	}
}
