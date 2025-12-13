package discussions

import "time"

const (
	BotUserID = 0
)

type DiscussionDraft struct {
	GiveawayID int64
	UserID     int64
	Text       string
}

type Discussion struct {
	DiscussionDraft

	ID                int64
	TelegramMessageID int64
	CreatedAt         time.Time
	UpdatedAt         time.Time
}
