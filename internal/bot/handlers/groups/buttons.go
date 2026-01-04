package groups

import (
	"github.com/capcom6/lucky-pick-tg-bot/internal/bot/handlers/settings"
	"github.com/go-telegram/bot/models"
)

// managementKeyboard creates keyboard for group management options including settings access.
func managementKeyboard(groupID int64) *models.InlineKeyboardMarkup {
	return &models.InlineKeyboardMarkup{
		InlineKeyboard: [][]models.InlineKeyboardButton{
			{
				{
					Text:         "⚙️ Manage Settings",
					CallbackData: settings.NewGroupSettingsData(groupID),
				},
			},
			{
				{
					Text:         "🔙 Back to Groups",
					CallbackData: "groups:back",
				},
			},
		},
	}
}
