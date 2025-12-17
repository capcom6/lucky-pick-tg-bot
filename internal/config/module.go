package config

import (
	"github.com/capcom6/lucky-pick-tg-bot/internal/discussions"
	"github.com/capcom6/lucky-pick-tg-bot/pkg/gotelegrambotfx"
	"github.com/go-core-fx/cachefx"
	"github.com/go-core-fx/fiberfx"
	"github.com/go-core-fx/logger"
	"github.com/go-core-fx/openrouterfx"
	"github.com/go-core-fx/sqlfx"
	"go.uber.org/fx"
)

func Module() fx.Option {
	return fx.Module(
		"config",
		logger.WithNamedLogger("config"),
		fx.Provide(
			New,
			fx.Private,
		),
		fx.Provide(
			func(cfg Config) fiberfx.Config {
				return fiberfx.Config{
					Address:     cfg.HTTP.Address,
					ProxyHeader: cfg.HTTP.ProxyHeader,
					Proxies:     cfg.HTTP.Proxies,
				}
			},
		),
		fx.Provide(
			func(cfg Config) sqlfx.Config {
				return sqlfx.Config{
					URL:             cfg.Database.URL,
					ConnMaxIdleTime: cfg.Database.ConnMaxIdleTime,
					ConnMaxLifetime: cfg.Database.ConnMaxLifetime,
					MaxOpenConns:    cfg.Database.MaxOpenConns,
					MaxIdleConns:    cfg.Database.MaxIdleConns,
				}
			},
		),
		fx.Provide(
			func(cfg Config) gotelegrambotfx.Config {
				return gotelegrambotfx.Config{
					Token: cfg.Telegram.Token,
				}
			},
		),
		fx.Provide(
			func(cfg Config) openrouterfx.Config {
				return openrouterfx.Config{
					Token:   cfg.OpenRouter.Token,
					AppName: "LuckyPick Bot",
					AppURL:  "https://lucky-pick.ru",
				}
			},
		),
		fx.Provide(
			func(cfg Config) discussions.Config {
				return discussions.Config{
					LLMModel: cfg.Discussions.LLMModel,
				}
			},
		),
		fx.Provide(
			func(cfg Config) cachefx.Config {
				return cachefx.Config{
					URL: cfg.Cache.URL,
				}
			},
		),
	)
}
