package adaptor

import (
	"context"

	"github.com/capcom6/lucky-pick-tg-bot/internal/bot/middlewares/state"
	"github.com/capcom6/lucky-pick-tg-bot/internal/bot/middlewares/user"
	"github.com/capcom6/lucky-pick-tg-bot/internal/fsm"
	"github.com/capcom6/lucky-pick-tg-bot/internal/users"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

type Context struct {
	context.Context
}

func (c *Context) User() (*users.User, error) {
	return user.FromContext(c.Context)
}

func (c *Context) State() (*fsm.State, error) {
	return state.FromContext(c.Context)
}

func New(h func(*Context, *models.Update)) bot.HandlerFunc {
	return func(ctx context.Context, b *bot.Bot, update *models.Update) {
		h(&Context{Context: ctx}, update)
	}
}
