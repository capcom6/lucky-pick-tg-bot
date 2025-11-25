package db

import (
	"embed"
	"fmt"
	"io/fs"

	"github.com/go-core-fx/goosefx"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

func NewMigrationsStorage() (goosefx.Storage, error) {
	sub, err := fs.Sub(migrationsFS, "migrations")
	if err != nil {
		return nil, fmt.Errorf("embed migrations fs: %w", err)
	}
	return goosefx.Storage(sub), nil
}
