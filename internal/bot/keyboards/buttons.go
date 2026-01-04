package keyboards

import "github.com/go-telegram/bot/models"

// Navigation button builders

// BackButton creates a standard back button
func BackButton(callbackData string) models.InlineKeyboardButton {
	return models.InlineKeyboardButton{
		Text:         "🔙 Back",
		CallbackData: callbackData,
	}
}

// CancelButton creates a standard cancel button
func CancelButton() models.InlineKeyboardButton {
	return models.InlineKeyboardButton{
		Text:         "❌ Cancel",
		CallbackData: "settings:cancel",
	}
}

// HomeButton creates a button to return to main menu
func HomeButton() models.InlineKeyboardButton {
	return models.InlineKeyboardButton{
		Text:         "🏠 Main Menu",
		CallbackData: "groups:back",
	}
}

// SaveButton creates a save confirmation button
func SaveButton(callbackData string) models.InlineKeyboardButton {
	return models.InlineKeyboardButton{
		Text:         "💾 Save",
		CallbackData: callbackData,
	}
}

// EditButton creates an edit button
func EditButton(callbackData string) models.InlineKeyboardButton {
	return models.InlineKeyboardButton{
		Text:         "✏️ Edit",
		CallbackData: callbackData,
	}
}

// Common navigation row builders

// SingleBackRow creates a row with a single back button
func SingleBackRow(callbackData string) []models.InlineKeyboardButton {
	return []models.InlineKeyboardButton{BackButton(callbackData)}
}

// BackAndHomeRow creates a row with back and home buttons
func BackAndHomeRow(backCallback string) []models.InlineKeyboardButton {
	return []models.InlineKeyboardButton{
		BackButton(backCallback),
		HomeButton(),
	}
}

// SaveEditCancelRow creates a row with save, edit again, and cancel buttons
func SaveEditCancelRow(saveCallback, editCallback string) []models.InlineKeyboardButton {
	return []models.InlineKeyboardButton{
		SaveButton(saveCallback),
		EditButton(editCallback),
		CancelButton(),
	}
}
