package blueprint

import (
	"github.com/reddec/dreaming-bard/internal/dbo"
	"github.com/reddec/dreaming-bard/internal/utils/xsync"
)

type indexParams struct {
	Blueprint        dbo.Blueprint
	Pages            []dbo.ListBlueprintPagesRow
	Steps            []dbo.BlueprintStep
	EditStep         int64
	AvailableContext []dbo.Context
	LinkedContext    []dbo.Context
	Roles            []dbo.Role
	Chats            []dbo.Chat
	Tasks            []*xsync.Task[dbo.BlueprintStep]
	PlanningTasks    []*xsync.Task[dbo.Blueprint]
}

func (ip indexParams) AvailableFacts() []dbo.Context {
	var result []dbo.Context
	for _, c := range ip.AvailableContext {
		if !c.Archived {
			result = append(result, c)
		}
	}
	return result
}

func (ip indexParams) IsActiveStep(step int64) bool {
	for _, t := range ip.Tasks {
		if t.State().ID == step {
			return t.IsRunning()
		}
	}
	return false
}

func (ip indexParams) IsPlanning() bool {
	for _, t := range ip.PlanningTasks {
		if t.State().ID == ip.Blueprint.ID {
			return t.IsRunning()
		}
	}
	return false
}

type listParams struct {
	Blueprints []dbo.Blueprint
	ShowHelp   bool
}

type wizardParams struct {
	Pages []dbo.Page
}
