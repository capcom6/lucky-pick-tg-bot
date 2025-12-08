package users

import (
	"context"
	"database/sql"
	"errors"
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

// CreateOrUpdate creates a new user or updates existing one.
func (r *Repository) CreateOrUpdate(ctx context.Context, user *UserModel) (bool, error) {
	created := false

	err := r.db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
		existing := new(UserModel)
		if err := tx.NewSelect().
			For("UPDATE").
			Model(existing).
			Where("telegram_user_id = ?", user.TelegramUserID).
			Scan(ctx); err != nil && !errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("failed to get user by Telegram ID: %w", err)
		}

		_, err := tx.NewInsert().
			Model(user).
			On("DUPLICATE KEY UPDATE").
			Returning("*").
			Exec(ctx)

		if err != nil {
			return fmt.Errorf("failed to insert user: %w", err)
		}

		if existing.ID == 0 {
			created = true
		}

		return nil
	})

	if err != nil {
		return false, fmt.Errorf("failed to create or update user: %w", err)
	}

	return created, nil
}
