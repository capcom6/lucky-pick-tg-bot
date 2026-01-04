package adaptor

import (
	"context"
	"fmt"

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
	u, err := user.FromContext(c.Context)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return u, nil
}

func (c *Context) State() (*fsm.State, error) {
	s, err := state.FromContext(c.Context)
	if err != nil {
		return nil, fmt.Errorf("failed to get state: %w", err)
	}

	return s, nil
}

func New(h func(*Context, *models.Update)) bot.HandlerFunc {
	return func(ctx context.Context, _ *bot.Bot, update *models.Update) {
		h(&Context{Context: ctx}, update)
	}
}
