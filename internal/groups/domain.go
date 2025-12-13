package groups

import "time"

// GroupDraft represents a Telegram group in the system.
type GroupDraft struct {
	TelegramID           int64
	Title                string
	DiscussionsThreshold int
}

type Group struct {
	GroupDraft

	ID        int64
	CreatedAt time.Time
	UpdatedAt time.Time
}

type Admin struct {
	UserID int64
}
