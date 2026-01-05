package keyboards

import "github.com/go-telegram/bot/models"

// Navigation button builders

// BackButton creates a standard back button.
func BackButton(callbackData string) models.InlineKeyboardButton {
	return models.InlineKeyboardButton{
		Text:         "🔙 Back",
		CallbackData: callbackData,
	}
}

// HomeButton creates a button to return to main menu.
func HomeButton(callbackData string) models.InlineKeyboardButton {
	return models.InlineKeyboardButton{
		Text:         "🏠 Main Menu",
		CallbackData: callbackData,
	}
}

// Common navigation row builders

// SingleBackRow creates a row with a single back button.
func SingleBackRow(callbackData string) []models.InlineKeyboardButton {
	return []models.InlineKeyboardButton{BackButton(callbackData)}
}

// BackAndHomeRow creates a row with back and home buttons.
func BackAndHomeRow(backCallback string, homeCallback string) []models.InlineKeyboardButton {
	return []models.InlineKeyboardButton{
		BackButton(backCallback),
		HomeButton(homeCallback),
	}
}
