package settings

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/capcom6/lucky-pick-tg-bot/internal/settings"
	"github.com/go-telegram/bot/models"
)

const (
	settingsCallbackCategory = "settings:category:"
	settingsCallbackSetting  = "settings:setting:"
)

// categoriesKeyboard creates keyboard for the main settings menu showing categories
func categoriesKeyboard(categories []string) *models.InlineKeyboardMarkup {
	var keyboard [][]models.InlineKeyboardButton

	// Add category buttons
	for _, category := range categories {
		displayName := getCategoryDisplayName(category)
		keyboard = append(keyboard, []models.InlineKeyboardButton{
			{
				Text:         displayName,
				CallbackData: settingsCallbackCategory + category,
			},
		})
	}

	// Add navigation buttons
	keyboard = append(keyboard, []models.InlineKeyboardButton{
		{
			Text:         "🔙 Back to Groups",
			CallbackData: "groups:back",
		},
	})

	return &models.InlineKeyboardMarkup{
		InlineKeyboard: keyboard,
	}
}

// settingsKeyboard creates keyboard for listing settings in a category
func settingsKeyboard(
	groupID int64,
	settingsList []settings.SettingDefinition,
	currentValues map[string]interface{},
) *models.InlineKeyboardMarkup {
	var keyboard [][]models.InlineKeyboardButton

	// Add setting buttons with current values
	for _, setting := range settingsList {
		currentValue := currentValues[setting.Key]
		displayValue := formatSettingValue(setting.Type, currentValue)

		buttonText := fmt.Sprintf("%s: %s", setting.Label, displayValue)

		keyboard = append(keyboard, []models.InlineKeyboardButton{
			{
				Text:         buttonText,
				CallbackData: settingsCallbackSetting + setting.Key,
			},
		})
	}

	// Add navigation buttons
	keyboard = append(keyboard, []models.InlineKeyboardButton{
		{
			Text:         "🔙 Back to Categories",
			CallbackData: callbackListPrefix + strconv.FormatInt(groupID, 10),
		},
		{
			Text:         "🏠 Main Menu",
			CallbackData: "groups:back",
		},
	})

	return &models.InlineKeyboardMarkup{
		InlineKeyboard: keyboard,
	}
}

// editKeyboard creates keyboard for editing a specific setting
func editKeyboard(
	setting settings.SettingDefinition,
	currentValue interface{},
) *models.InlineKeyboardMarkup {
	var keyboard [][]models.InlineKeyboardButton

	switch setting.Type {
	case settings.Boolean:
		// For boolean, show toggle buttons
		currentBool, _ := currentValue.(bool)
		keyboard = buildBooleanToggleKeyboard(currentBool)
	default:
		// For other types, show edit button
		keyboard = [][]models.InlineKeyboardButton{}
	}

	// Add navigation buttons
	// keyboard = append(keyboard, []models.InlineKeyboardButton{
	// 	{
	// 		Text:         "🔙 Back to Settings",
	// 		CallbackData: "settings:back_to_list",
	// 	},
	// 	{
	// 		Text:         "🏠 Categories",
	// 		CallbackData: "settings:back_to_categories",
	// 	},
	// })

	return &models.InlineKeyboardMarkup{
		InlineKeyboard: keyboard,
	}
}

// Helper functions

func getCategoryDisplayName(category string) string {
	categoryMap := map[string]string{
		"discussions":   "💬 Discussions",
		"giveaways":     "🎯 Giveaways",
		"moderation":    "🛡️ Moderation",
		"notifications": "🔔 Notifications",
		"general":       "⚙️ General",
	}

	if display, exists := categoryMap[category]; exists {
		return display
	}

	return strings.Title(category)
}

func buildBooleanToggleKeyboard(_ bool) [][]models.InlineKeyboardButton {
	return [][]models.InlineKeyboardButton{
		{
			{
				Text:         "✅ True",
				CallbackData: settingInputBoolean + "true",
			},
			{
				Text:         "❌ False",
				CallbackData: settingInputBoolean + "false",
			},
		},
	}
}

// Utility functions for keyboard patterns and value formatting

// formatSettingValue formats a setting value for display in keyboard buttons
func formatSettingValue(settingType settings.SettingType, value interface{}) string {
	return formatSettingDisplayValue(settingType, valueToString(value))
}

// formatSettingDisplayValue formats a string value for display based on setting type
func formatSettingDisplayValue(settingType settings.SettingType, value string) string {
	switch settingType {
	case settings.Boolean:
		if value == "true" {
			return "✅ True"
		}
		return "❌ False"
	case settings.Duration:
		// Parse duration and format nicely
		if dur, err := settings.ParseDuration(value); err == nil {
			return dur.String()
		}
		return value
	default:
		return value
	}
}

// valueToString converts various value types to string representation
func valueToString(value interface{}) string {
	if value == nil {
		return ""
	}

	switch v := value.(type) {
	case string:
		return v
	case bool:
		if v {
			return "true"
		}
		return "false"
	case int, int32, int64:
		return fmt.Sprintf("%d", v)
	case float32, float64:
		return fmt.Sprintf("%g", v)
	case time.Duration:
		return v.String()
	case settings.DurationValue:
		return v.String()
	default:
		return fmt.Sprintf("%v", v)
	}
}
