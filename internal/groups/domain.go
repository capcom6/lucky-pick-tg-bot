package groups

import (
	"time"

	"github.com/samber/lo"
)

// GroupDraft represents a Telegram group in the system.
type GroupDraft struct {
	TelegramID int64
	Title      string
}

type Group struct {
	GroupDraft

	ID int64

	CreatedAt time.Time
	UpdatedAt time.Time
}

func newGroup(model *GroupModel) *Group {
	return &Group{
		GroupDraft: GroupDraft{
			TelegramID: model.TelegramGroupID,
			Title:      model.Title,
		},
		ID:        model.ID,
		CreatedAt: model.CreatedAt,
		UpdatedAt: model.UpdatedAt,
	}
}

type GroupWithSettings struct {
	Group

	Settings map[string]string
}

func newGroupWithSettings(model *GroupModel) *GroupWithSettings {
	return &GroupWithSettings{
		Group: *newGroup(model),
		Settings: lo.SliceToMap(
			model.Settings,
			func(item *settingsModel) (string, string) { return item.Key, item.Value },
		),
	}
}

type Admin struct {
	UserID int64
}
