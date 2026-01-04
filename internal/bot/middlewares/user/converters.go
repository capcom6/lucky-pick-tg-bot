package user

import (
	"github.com/capcom6/lucky-pick-tg-bot/internal/users"
	"github.com/go-telegram/bot/models"
)

func ToDomain(user *models.User) users.UserIn {
	if user == nil {
		return users.UserIn{}
	}

	return users.UserIn{
		TelegramUserID: user.ID,
		Username:       user.Username,
		FirstName:      user.FirstName,
		LastName:       user.LastName,
	}
}
