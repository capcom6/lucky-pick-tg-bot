package discussions

import (
	"time"

	"github.com/capcom6/lucky-pick-tg-bot/internal/settings"
	"github.com/go-core-fx/logger"
	"go.uber.org/fx"
)

func Module() fx.Option {
	return fx.Module(
		"discussions",
		logger.WithNamedLogger("discussions"),
		fx.Provide(NewRepository, fx.Private),
		fx.Provide(NewLLM, fx.Private),
		fx.Provide(NewService),
		fx.Invoke(func(settingsSvc *settings.Service) {
			settingsSvc.RegisterDefinition(settings.SettingDefinition{
				Key:          "discussions.delay",
				Category:     "discussions",
				Label:        "Discussion Delay",
				Description:  "Time before a new discussion is started",
				Type:         settings.Duration,
				DefaultValue: settings.DurationValue{Duration: 6 * time.Hour},
				Validation: &settings.SettingValidation{
					MinValue: settings.Ptr(float64(3600)),      // Minimum 1 hour
					MaxValue: settings.Ptr(float64(24 * 3600)), // Maximum 24 hours
					Required: false,
				},
			})
		}),
	)
}
