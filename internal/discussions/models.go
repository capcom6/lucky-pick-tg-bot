package discussions

import (
	"time"

	"github.com/uptrace/bun"
)

type discussionModel struct {
	bun.BaseModel `bun:"table:giveaway_discussions,alias:gd"`

	ID                int64     `bun:"id,pk,autoincrement"`
	GiveawayID        int64     `bun:"giveaway_id,notnull"`
	UserID            int64     `bun:"user_id,nullzero"`
	TelegramMessageID int64     `bun:"telegram_message_id,nullzero"`
	Text              string    `bun:"text,notnull"`
	CreatedAt         time.Time `bun:"created_at,scanonly"`
	UpdatedAt         time.Time `bun:"updated_at,scanonly"`
}

func newDiscussionModel(draft DiscussionDraft) *discussionModel {
	//nolint:exhaustruct // partial constructor
	return &discussionModel{
		GiveawayID: draft.GiveawayID,
		UserID:     draft.UserID,
		Text:       draft.Text,
	}
}

func (d *discussionModel) toDiscussion() *Discussion {
	return &Discussion{
		DiscussionDraft: DiscussionDraft{
			GiveawayID: d.GiveawayID,
			UserID:     d.UserID,
			Text:       d.Text,
		},

		ID:                d.ID,
		TelegramMessageID: d.TelegramMessageID,
		CreatedAt:         d.CreatedAt,
		UpdatedAt:         d.UpdatedAt,
	}
}
