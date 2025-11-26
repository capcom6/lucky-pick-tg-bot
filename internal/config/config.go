package config

import (
	"fmt"
	"time"

	"github.com/go-core-fx/config"
	"go.uber.org/zap"
)

type httpConfig struct {
	Address     string   `koanf:"address"`
	ProxyHeader string   `koanf:"proxy_header"`
	Proxies     []string `koanf:"proxies"`
}

type telegramConfig struct {
	Token string `koanf:"token"`
}

type databaseConfig struct {
	URL             string        `koanf:"url"`
	ConnMaxIdleTime time.Duration `koanf:"conn_max_idle_time"`
	ConnMaxLifetime time.Duration `koanf:"conn_max_lifetime"`
	MaxOpenConns    int           `koanf:"max_open_conns"`
	MaxIdleConns    int           `koanf:"max_idle_conns"`
}

type Config struct {
	HTTP     httpConfig     `koanf:"http"`
	Telegram telegramConfig `koanf:"telegram"`
	Database databaseConfig `koanf:"database"`
}

func New(logger *zap.Logger) (Config, error) {
	//nolint:mnd // default values
	cfg := Config{
		HTTP: httpConfig{
			Address:     "127.0.0.1:3000",
			ProxyHeader: "X-Forwarded-For",
			Proxies:     []string{},
		},
		Telegram: telegramConfig{
			Token: "",
		},
		Database: databaseConfig{
			URL:             "mariadb://lucky-pick:lucky-pick@127.0.0.1:3306/lucky-pick?charset=utf8mb4&parseTime=True&loc=UTC",
			ConnMaxIdleTime: 10 * time.Minute,
			ConnMaxLifetime: 1 * time.Hour,
			MaxOpenConns:    25,
			MaxIdleConns:    5,
		},
	}

	if err := config.Load(&cfg); err != nil {
		logger.Error("failed to load config", zap.Error(err))
		return cfg, fmt.Errorf("load config: %w", err)
	}

	logger.Debug("config loaded successfully")

	return cfg, nil
}
