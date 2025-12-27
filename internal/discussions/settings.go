package discussions

import (
	"fmt"
	"strconv"
	"time"
)

type Settings struct {
	Delay time.Duration
}

func DefaultSettings() Settings {
	return Settings{
		Delay: 0,
	}
}

func NewSettings(settings map[string]string) (Settings, error) {
	s := DefaultSettings()

	if d := settings["discussions.delay"]; d != "" {
		delay, err := strconv.Atoi(d)
		if err != nil {
			return s, fmt.Errorf("failed to parse delay: %w", err)
		}
		s.Delay = time.Duration(delay) * time.Hour
	}

	return s, nil
}
