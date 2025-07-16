package pages

import (
	"github.com/reddec/dreaming-bard/internal/dbo"
)

type indexParams struct {
	Page  dbo.Page
	Roles []dbo.Role
}

type importParams struct {
}

type newParams struct {
	Prefill string
}

type listParams struct {
	Pages    []dbo.Page
	ShowHelp bool
}

type editParams struct {
	Page dbo.Page
}

type epubParams struct {
}
