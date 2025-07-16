package chats

import (
	"maps"
	"slices"

	"github.com/reddec/dreaming-bard/internal/common"
	"github.com/reddec/dreaming-bard/internal/dbo"
)

type indexParams struct {
	UID           int64
	Role          dbo.Role
	Chat          dbo.Chat
	History       []dbo.Message
	IsBusy        bool
	Facts         []dbo.Context
	EditMessageID int64
	Pages         []dbo.Page
}

type listParams struct {
	Threads   []dbo.ListChatsRow
	BusyChats []int64
	ShowHelp  bool
}

func (lp listParams) IsBusy(chatID int64) bool {
	for _, c := range lp.BusyChats {
		if c == chatID {
			return true
		}
	}
	return false
}

type roleGroup struct {
	Order   int
	Purpose common.Purpose
	Roles   []dbo.Role
}

type wizardParams struct {
	Prompt dbo.Prompt
	Facts  []dbo.Context
	Roles  []dbo.Role
}

func (wp wizardParams) Group() []*roleGroup {
	var extra = len(common.PurposeValues()) + 1
	var groupByPurpose = make(map[common.Purpose]*roleGroup)
	for _, role := range wp.Roles {
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
