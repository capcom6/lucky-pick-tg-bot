package filter

import (
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

func And(filters ...bot.MatchFunc) bot.MatchFunc {
	return func(update *models.Update) bool {
		for _, f := range filters {
			if !f(update) {
				return false
			}
		}
		return true
	}
}
