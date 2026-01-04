package settings

import (
	"fmt"
	"strconv"

	"github.com/capcom6/lucky-pick-tg-bot/internal/bot/keyboards"
	"github.com/capcom6/lucky-pick-tg-bot/internal/settings"
	"github.com/go-telegram/bot/models"
)

const (
	callbackGroupPrefix    = "settings:group:"
	callbackCategoryPrefix = "settings:category:"
	callbackSettingPrefix  = "settings:setting:"

	callbackInputBooleanPrefix = "settings:input:boolean:"
)

// categoriesKeyboard creates keyboard for the main settings menu showing categories.
func categoriesKeyboard(categories []string) *models.InlineKeyboardMarkup {
	var keyboard [][]models.InlineKeyboardButton

	// Add category buttons
	for _, category := range categories {
		keyboard = append(keyboard, []models.InlineKeyboardButton{
			{
				Text:         category,
				CallbackData: callbackCategoryPrefix + category,
			},
		})
	}

	// Add navigation buttons

	keyboard = append(keyboard, keyboards.SingleBackRow("groups:back"))

	return &models.InlineKeyboardMarkup{
		InlineKeyboard: keyboard,
	}
}

// settingsKeyboard creates keyboard for listing settings in a category.
func settingsKeyboard(
	groupID int64,
	settingsList []settings.SettingDefinition,
	currentValues map[string]string,
) *models.InlineKeyboardMarkup {
	var keyboard [][]models.InlineKeyboardButton

	// Add setting buttons with current values
	for _, setting := range settingsList {
		currentValue := currentValues[setting.Key]
		displayValue := setting.Format(currentValue)

		buttonText := fmt.Sprintf("%s: %s", setting.Label, displayValue)

		keyboard = append(keyboard, []models.InlineKeyboardButton{
			{
				Text:         buttonText,
				CallbackData: callbackSettingPrefix + setting.Key,
			},
		})
	}

	// Add navigation buttons
	keyboard = append(
		keyboard,
		keyboards.BackAndHomeRow(callbackGroupPrefix+strconv.FormatInt(groupID, 10), "groups:back"),
	)

	return &models.InlineKeyboardMarkup{
		InlineKeyboard: keyboard,
	}
}

// editKeyboard creates keyboard for editing a specific setting.
func editKeyboard(
	setting settings.SettingDefinition,
	currentValue string,
) *models.InlineKeyboardMarkup {
	var keyboard [][]models.InlineKeyboardButton

	switch setting.Type {
	case settings.Boolean:
		// For boolean, show toggle buttons
		keyboard = buildBooleanToggleKeyboard(currentValue)
	case settings.Duration, settings.Number, settings.Text:
		// For other types, show edit button
		keyboard = [][]models.InlineKeyboardButton{}
	}

	return &models.InlineKeyboardMarkup{
		InlineKeyboard: keyboard,
	}
}

func buildBooleanToggleKeyboard(currentBool string) [][]models.InlineKeyboardButton {
	trueText := "True"
	falseText := "False"
	if v, err := strconv.ParseBool(currentBool); err == nil && v {
		trueText = "✅ True (current)"
	} else {
		falseText = "❌ False (current)"
	}
	return [][]models.InlineKeyboardButton{
		{
			{
				Text:         trueText,
				CallbackData: callbackInputBooleanPrefix + "true",
			},
			{
				Text:         falseText,
				CallbackData: callbackInputBooleanPrefix + "false",
			},
		},
	}
}
