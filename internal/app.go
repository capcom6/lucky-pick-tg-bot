package internal

import (
	"context"

	"github.com/capcom6/lucky-pick-tg-bot/internal/actions"
	"github.com/capcom6/lucky-pick-tg-bot/internal/bot"
	"github.com/capcom6/lucky-pick-tg-bot/internal/config"
	"github.com/capcom6/lucky-pick-tg-bot/internal/db"
	"github.com/capcom6/lucky-pick-tg-bot/internal/discussions"
	"github.com/capcom6/lucky-pick-tg-bot/internal/fsm"
	"github.com/capcom6/lucky-pick-tg-bot/internal/giveaways"
	"github.com/capcom6/lucky-pick-tg-bot/internal/groups"
	"github.com/capcom6/lucky-pick-tg-bot/internal/llm"
	"github.com/capcom6/lucky-pick-tg-bot/internal/scheduler"
	"github.com/capcom6/lucky-pick-tg-bot/internal/server"
	"github.com/capcom6/lucky-pick-tg-bot/internal/users"
	"github.com/capcom6/lucky-pick-tg-bot/pkg/gotelegrambotfx"
	"github.com/capcom6/lucky-pick-tg-bot/pkg/openrouterfx"
	"github.com/go-core-fx/bunfx"
	"github.com/go-core-fx/cachefx"
	"github.com/go-core-fx/fiberfx"
	"github.com/go-core-fx/goosefx"
	"github.com/go-core-fx/logger"
	"github.com/go-core-fx/sqlfx"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

func Run() {
	fx.New(
		logger.Module(),
		logger.WithFxDefaultLogger(),
		config.Module(),
		sqlfx.Module(),
		goosefx.Module(),
		bunfx.Module(),
		fiberfx.Module(),
		cachefx.Module(),
		gotelegrambotfx.Module(),
		openrouterfx.Module(),
		//
		db.Module(),
		server.Module(),
		bot.Module(),
		scheduler.Module(),
		llm.Module(),
		fsm.Module(),
		//
		users.Module(),
		giveaways.Module(),
		groups.Module(),
		actions.Module(),
		discussions.Module(),
		//
		fx.Invoke(func(lc fx.Lifecycle, logger *zap.Logger) {
			lc.Append(fx.Hook{
				OnStart: func(_ context.Context) error {
					logger.Info("app started")
					return nil
				},
				OnStop: func(_ context.Context) error {
					logger.Info("app stopped")
					return nil
				},
			})
		}),
	).Run()
}
