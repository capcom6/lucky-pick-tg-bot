package extractors

import "github.com/go-telegram/bot/models"

func UserID(update *models.Update) int64 {
	if update == nil {
		return 0
	}

	if update.Message != nil && update.Message.From != nil {
		return update.Message.From.ID
	}
	if update.CallbackQuery != nil {
		return update.CallbackQuery.From.ID
	}
	return 0
}

func ChatID(update *models.Update) int64 {
	if update == nil {
		return 0
	}

	if update.Message != nil {
		return update.Message.Chat.ID
	}
	if update.MyChatMember != nil {
		return update.MyChatMember.Chat.ID
	}
	return 0
}

func From(update *models.Update) int64 {
	if userID := ChatID(update); userID != 0 {
		return userID
	}
	return UserID(update)
}
