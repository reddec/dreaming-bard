package contexts

import (
	"time"

	"github.com/reddec/dreaming-bard/internal/dbo"
)

type factMeta struct {
	ID       int64     `yaml:"id,omitempty"`
	Category string    `yaml:"category,omitempty"`
	Title    string    `yaml:"title,omitempty"`
	Created  time.Time `yaml:"created,omitempty"`
	Updated  time.Time `yaml:"updated,omitempty"`
}

type editParams struct {
	Fact dbo.Context
}

type listParams struct {
	Facts      []dbo.Context
	Categories []string
	Category   string
	ShowHelp   bool
}

func (lp listParams) ActiveFacts() []dbo.Context {
	var result []dbo.Context
	for _, fact := range lp.Facts {
		if !fact.Archived {
			result = append(result, fact)
		}
	}
	return result
}

func (lp listParams) ArchivedFacts() []dbo.Context {
	var result []dbo.Context
	for _, fact := range lp.Facts {
		if fact.Archived {
			result = append(result, fact)
		}
	}
	return result
}

type importParams struct {
}

type newParams struct {
	Facts      []dbo.Context
	Categories []string
	Prefill    map[string]string
	ManualOnly bool
}

type indexParams struct {
	Fact dbo.Context
}
