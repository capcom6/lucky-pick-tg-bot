package db

import (
	"embed"
	"fmt"
	"io/fs"

	"github.com/go-core-fx/goosefx"
	"go.uber.org/zap"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

func NewMigrationsStorage(logger *zap.Logger) (goosefx.Storage, error) {
	logger.Debug("initializing migration storage")

	sub, err := fs.Sub(migrationsFS, "migrations")
	if err != nil {
		return nil, fmt.Errorf("embed migrations fs: %w", err)
	}

	logger.Debug("migration storage initialized successfully")

	return goosefx.Storage(sub), nil
}
