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

		if existing.ID == 0 {
			created = true
			if err := r.insert(ctx, tx, user); err != nil {
				return err
			}
		} else {
			user.ID = existing.ID
			if err := r.update(ctx, tx, user); err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		return false, fmt.Errorf("failed to create or update user: %w", err)
	}

	return created, nil
}

func (r *Repository) insert(ctx context.Context, tx bun.Tx, user *UserModel) error {
	_, err := tx.NewInsert().
		Model(user).
		Exec(ctx)

	if err != nil {
		return fmt.Errorf("failed to insert user: %w", err)
	}

	return nil
}

func (r *Repository) update(ctx context.Context, tx bun.Tx, user *UserModel) error {
	_, err := tx.NewUpdate().
		Model(user).
		OmitZero().
		WherePK().
		Exec(ctx)

	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	return nil
}

func (r *Repository) GetByID(ctx context.Context, userID int64) (*UserModel, error) {
	user := new(UserModel)
	if err := r.db.NewSelect().
		Model(user).
		Where("id = ?", userID).
		Scan(ctx); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}

		return nil, fmt.Errorf("failed to get user by ID: %w", err)
	}

	return user, nil
}

func (r *Repository) SetActive(ctx context.Context, userID int64, isActive bool) error {
	if _, err := r.db.NewUpdate().
		Model((*UserModel)(nil)).
		Set("is_active = ?", isActive).
		Where("id = ?", userID).
		Exec(ctx); err != nil {
		return fmt.Errorf("failed to update user activity: %w", err)
	}

	return nil
}
