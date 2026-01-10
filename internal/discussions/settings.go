package discussions

import (
	"fmt"
	"time"

	"github.com/capcom6/lucky-pick-tg-bot/internal/settings"
)

type Settings struct {
	Delay time.Duration
}

func NewSettings(dict map[string]string) (Settings, error) {
	s := DefaultSettings()

	if d := dict["discussions.delay"]; d != "" {
		duration, err := settings.ParseDuration(d)
		if err != nil {
			return s, fmt.Errorf("failed to parse delay: %w", err)
		}
		s.Delay = duration.Duration
	}

	return s, nil
}

func DefaultSettings() Settings {
	//nolint:mnd //default values
	return Settings{
		Delay: time.Hour * 6,
	}
}

func SettingDefinitions() []settings.SettingDefinition {
	//nolint:exhaustruct,mnd //default values
	return []settings.SettingDefinition{
		{
			Key:          "discussions.delay",
			Category:     "💬 Discussions",
			Label:        "Discussion Delay",
			Description:  "Time before a new discussion is started",
			Type:         settings.Duration,
			DefaultValue: "06:00:00",
			Validation: &settings.SettingValidation{
				MinValue: settings.Ptr(float64(time.Hour.Seconds())),      // Minimum 1 hour
				MaxValue: settings.Ptr(float64(24 * time.Hour.Seconds())), // Maximum 24 hours
				Required: false,
			},
		},
	}
}
