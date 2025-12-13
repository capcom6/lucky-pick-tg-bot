package openrouterfx

import (
	"fmt"

	"github.com/revrost/go-openrouter"
)

func New(config Config) (*openrouter.Client, error) {
	if config.Token == "" {
		return nil, fmt.Errorf("%w: token is empty", ErrInvalidConfig)
	}

	var options []openrouter.Option
	if config.AppName != "" {
		options = append(options, openrouter.WithXTitle(config.AppName))
	}
	if config.AppURL != "" {
		options = append(options, openrouter.WithHTTPReferer(config.AppURL))
	}

	return openrouter.NewClient(config.Token, options...), nil
}
