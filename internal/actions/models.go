package actions

import (
	"time"

	"github.com/uptrace/bun"
)

type Entry struct {
	bun.BaseModel `bun:"table:action_logs,alias:al"`

	ID          uint64    `bun:"id,pk,autoincrement"`
	GiveawayID  *int64    `bun:"giveaway_id,nullzero"`
	UserID      *int64    `bun:"user_id,nullzero"`
	ActionType  string    `bun:"action_type,notnull"`
	Description string    `bun:"description,notnull"`
	CreatedAt   time.Time `bun:"created_at,scanonly"`
}

func NewEntry(
	giveawayID *int64,
	userID *int64,
	actionType string,
	description string,
) *Entry {
	//nolint:exhaustruct // partial constructor
	return &Entry{
		GiveawayID:  giveawayID,
		UserID:      userID,
		ActionType:  actionType,
		Description: description,
	}
}
