package groups

// Group represents a Telegram group in the system.
type Group struct {
	TelegramID int64
	Title      string
}

type Admin struct {
	UserID int64
}
