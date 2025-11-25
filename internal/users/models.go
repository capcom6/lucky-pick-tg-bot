package users

import (
	"time"

	"github.com/uptrace/bun"
)

// UserModel represents a Telegram user in the system.
type UserModel struct {
	bun.BaseModel `bun:"table:users,alias:u"`

	ID             int64     `bun:"id,pk,scanonly"`
	TelegramUserID int64     `bun:"telegram_user_id"`
	Username       string    `bun:"username,nullzero"`
	FirstName      string    `bun:"first_name"`
	LastName       string    `bun:"last_name,nullzero"`
	RegisteredAt   time.Time `bun:"registered_at,scanonly"`
	IsActive       bool      `bun:"is_active"`
}

func NewUserModel(telegramUserID int64, username, firstName, lastName string) *UserModel {
	//nolint:exhaustruct // partial constructor
	return &UserModel{
		TelegramUserID: telegramUserID,
		Username:       username,
		FirstName:      firstName,
		LastName:       lastName,
		IsActive:       true,
	}
}
