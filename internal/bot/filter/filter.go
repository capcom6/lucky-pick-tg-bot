package filter

import (
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

// And combines multiple bot.MatchFunc predicates with logical AND.
// It returns a MatchFunc that evaluates to true only if all provided filters return true.
// Evaluation short-circuits on the first filter that returns false.
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
