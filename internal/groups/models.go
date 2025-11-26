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

	IsActive bool `bun:"is_active"`

	CreatedAt time.Time `bun:"created_at,scanonly"`
	UpdatedAt time.Time `bun:"updated_at,scanonly"`
}

func NewGroupModel(telegramID int64, title string) *GroupModel {
	//nolint:exhaustruct // partial constructor
	return &GroupModel{
		TelegramGroupID: telegramID,
		Title:           title,
		IsActive:        true,
	}
}
