package prompts

import (
	"github.com/reddec/dreaming-bard/internal/dbo"
)

type indexParams struct {
	Prompt dbo.Prompt
}

type editParams struct {
	Prompt dbo.Prompt
	Roles  []dbo.Role
}

type newParams struct {
	Roles []dbo.Role
}

type listParams struct {
	Prompts  []dbo.ListPromptsRow
	ShowHelp bool
}

func (lp listParams) Pinned() []dbo.ListPromptsRow {
	var result []dbo.ListPromptsRow
	for _, p := range lp.Prompts {
		if p.Prompt.PinnedAt != nil {
			result = append(result, p)
		}
	}
	return result
}

func (lp listParams) Unpinned() []dbo.ListPromptsRow {
	var result []dbo.ListPromptsRow
	for _, p := range lp.Prompts {
		if p.Prompt.PinnedAt == nil {
			result = append(result, p)
		}
	}
	return result
}

type importParams struct {
	Roles       []dbo.Role
	FirstWriter int64
}
