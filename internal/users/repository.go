package users

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

// CreateOrUpdate creates a new user or updates existing one.
func (r *Repository) CreateOrUpdate(ctx context.Context, user *UserModel) error {
	_, err := r.db.NewInsert().
		Model(user).
		On("DUPLICATE KEY UPDATE").
		Returning("*").
		Exec(ctx)

	if err != nil {
		return fmt.Errorf("failed to create or update user: %w", err)
	}

	return nil
}

// // GetByTelegramID retrieves a user by their Telegram user ID.
// func (r *Repository) GetByTelegramID(ctx context.Context, telegramID int64) (*UserModel, error) {
// 	user := new(UserModel)
// 	err := r.db.NewSelect().
// 		Model(&user).
// 		Where("telegram_user_id = ?", telegramID).
// 		Where("is_active = ?", true).
// 		Scan(ctx)

// 	if err != nil {
// 		if errors.Is(err, sql.ErrNoRows) {
// 			return nil, ErrNotFound
// 		}
// 		return nil, fmt.Errorf("failed to get user by telegram ID: %w", err)
// 	}

// 	return user, nil
// }

// // GetByID retrieves a user by their internal ID.
// func (r *Repository) GetByID(ctx context.Context, id int64) (*UserModel, error) {
// 	user := new(UserModel)
// 	err := r.db.NewSelect().
// 		Model(&user).
// 		Where("id = ?", id).
// 		Where("is_active = ?", true).
// 		Scan(ctx)

// 	if err != nil {
// 		if errors.Is(err, sql.ErrNoRows) {
// 			return nil, ErrNotFound
// 		}
// 		return nil, fmt.Errorf("failed to get user by ID: %w", err)
// 	}

// 	return user, nil
// }

// // IsUserRegistered checks if a user is registered and active.
// func (r *Repository) IsUserRegistered(ctx context.Context, telegramID int64) (bool, error) {
// 	var count int
// 	count, err := r.db.NewSelect().
// 		Model((*UserModel)(nil)).
// 		Where("telegram_user_id = ?", telegramID).
// 		Where("is_active = ?", true).
// 		Count(ctx)

// 	if err != nil {
// 		return false, fmt.Errorf("failed to check user registration: %w", err)
// 	}

// 	return count > 0, nil
// }

// // DeactivateUser deactivates a user (soft delete).
// func (r *Repository) DeactivateUser(ctx context.Context, telegramID int64) error {
// 	_, err := r.db.NewUpdate().
// 		Model((*UserModel)(nil)).
// 		Set("is_active = ?", false).
// 		Where("telegram_user_id = ?", telegramID).
// 		Exec(ctx)

// 	if err != nil {
// 		return fmt.Errorf("failed to deactivate user: %w", err)
// 	}

// 	return nil
// }
