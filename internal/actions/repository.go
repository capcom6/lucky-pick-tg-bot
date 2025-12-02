package actions

import (
	"context"
	"fmt"

	"github.com/uptrace/bun"
)

type Repository struct {
	db *bun.DB
}

func NewRepository(db *bun.DB) *Repository {
	return &Repository{
		db: db,
	}
}

func (r *Repository) LogAction(ctx context.Context, log *Entry) error {
	_, err := r.db.NewInsert().
		Model(log).
		Exec(ctx)

	if err != nil {
		return fmt.Errorf("failed to log action: %w", err)
	}

	return nil
}
