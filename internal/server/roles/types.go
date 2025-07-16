package roles

import (
	"maps"
	"slices"

	"github.com/reddec/dreaming-bard/internal/common"
	"github.com/reddec/dreaming-bard/internal/dbo"
)

type editParams struct {
	Role     dbo.Role
	Purposes []common.Purpose
}

type newParams struct {
	Content     string
	Description string
	Model       string
	Purpose     common.Purpose
	Purposes    []common.Purpose
}

type listParams struct {
	Roles    []dbo.Role
	ShowHelp bool
}

type roleGroup struct {
	Order   int
	Purpose common.Purpose
	Roles   []dbo.Role
}

func (lp listParams) Group() []*roleGroup {
	var extra = len(common.PurposeValues()) + 1
	var groupByPurpose = make(map[common.Purpose]*roleGroup)
	for _, role := range lp.Roles {
		if _, ok := groupByPurpose[role.Purpose]; !ok {
			order := slices.Index(common.PurposeValues(), role.Purpose)
			if order < 0 {
				order = extra
				extra++
			}
			groupByPurpose[role.Purpose] = &roleGroup{
				Purpose: role.Purpose,
				Order:   order,
			}
		}
		groupByPurpose[role.Purpose].Roles = append(groupByPurpose[role.Purpose].Roles, role)
	}
	return slices.SortedFunc(maps.Values(groupByPurpose), func(group *roleGroup, group2 *roleGroup) int {
		return group.Order - group2.Order
	})
}

type indexParams struct {
	Role dbo.Role
}

type importParams struct {
}
