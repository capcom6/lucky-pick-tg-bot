package groups

import (
	"time"

	"github.com/uptrace/bun"
)

type GroupModel struct {
	bun.BaseModel `bun:"table:groups,alias:g"`

	ID int64 `bun:"id,pk,autoincrement"`

	TelegramGroupID int64  `bun:"telegram_group_id"`
	Title           string `bun:"title"`

	IsActive             bool `bun:"is_active"`
	DiscussionsThreshold int  `bun:"discussions_period,nullzero"`

	CreatedAt time.Time `bun:"created_at,scanonly"`
	UpdatedAt time.Time `bun:"updated_at,scanonly"`

	Admins []*adminModel `bun:"ga,rel:has-many,join:id=group_id"`
}

func newGroupModel(draft *GroupDraft) *GroupModel {
	//nolint:exhaustruct // partial constructor
	return &GroupModel{
		TelegramGroupID:      draft.TelegramID,
		Title:                draft.Title,
		IsActive:             true,
		DiscussionsThreshold: draft.DiscussionsThreshold,
	}
}

type adminModel struct {
	bun.BaseModel `bun:"table:group_admins,alias:ga"`

	ID      int64 `bun:"id,pk,autoincrement"`
	GroupID int64 `bun:"group_id,notnull"`
	UserID  int64 `bun:"user_id,notnull"`
}

func newAdminModel(groupID, userID int64) *adminModel {
	//nolint:exhaustruct // partial constructor
	return &adminModel{
		GroupID: groupID,
		UserID:  userID,
	}
}
