package handlers

import (
	"github.com/capcom6/lucky-pick-tg-bot/internal/users"
	"github.com/go-telegram/bot/models"
)

func UserToDomain(user *models.User) users.UserIn {
	return users.UserIn{
		TelegramUserID: user.ID,
		Username:       user.Username,
		FirstName:      user.FirstName,
		LastName:       user.LastName,
	}
}
