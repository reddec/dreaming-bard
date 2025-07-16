package home

import (
	"github.com/reddec/dreaming-bard/internal/dbo"
	"github.com/reddec/dreaming-bard/internal/llm"
)

type indexParams struct {
	Config        llm.Provider
	PinnedPrompts []dbo.Prompt
	Pages         []dbo.Page
	Chats         []dbo.Chat
}
