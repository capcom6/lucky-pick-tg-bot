package keyboards

import (
	"strconv"

	"github.com/capcom6/lucky-pick-tg-bot/internal/bot/handlers/settings"
	"github.com/capcom6/lucky-pick-tg-bot/internal/groups"
	"github.com/go-telegram/bot/models"
	"github.com/samber/lo"
)

func GroupSelectionKeyboard(dataPrefix string, grps []groups.GroupWithSettings) *models.InlineKeyboardMarkup {
	return &models.InlineKeyboardMarkup{
		InlineKeyboard: lo.Map(grps, func(group groups.GroupWithSettings, _ int) []models.InlineKeyboardButton {
			return []models.InlineKeyboardButton{
				{
					Text:         group.Title,
					CallbackData: dataPrefix + strconv.FormatInt(group.ID, 10),
				},
			}
		}),
	}
}

// GroupManagementKeyboard creates keyboard for group management options including settings access
func GroupManagementKeyboard(groupID int64) *models.InlineKeyboardMarkup {
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
