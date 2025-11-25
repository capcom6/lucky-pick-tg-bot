package users

import "time"

// UserIn represents the data needed to create a new user.
type UserIn struct {
	TelegramUserID int64
	Username       string
	FirstName      string
	LastName       string
}

// User represents the data returned for a user.
type User struct {
	UserIn

	ID           int64
	RegisteredAt time.Time
	IsActive     bool
}
