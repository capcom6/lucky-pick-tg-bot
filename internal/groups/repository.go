package groups

import (
	"context"
	"fmt"

	"github.com/samber/lo"
	"github.com/uptrace/bun"
)

// Repository provides persistence operations for groups.
type Repository struct {
	db *bun.DB
}

// NewRepository creates a new instance of the repository.
func NewRepository(db *bun.DB) *Repository {
	return &Repository{db: db}
}

// CreateOrUpdate implements Repository.
func (r *Repository) CreateOrUpdate(ctx context.Context, group *Group, admins []Admin) error {
	err := r.db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
		model := newGroupModel(group.TelegramID, group.Title)

		if _, err := tx.NewInsert().
			Model(model).
			On("DUPLICATE KEY UPDATE").
			Returning("*").
			Exec(ctx); err != nil {
			return fmt.Errorf("failed to create or update group: %w", err)
		}

		if _, err := tx.NewDelete().
			Model((*adminModel)(nil)).
			Where("group_id = ?", model.ID).
			Exec(ctx); err != nil {
			return fmt.Errorf("failed to delete group admins: %w", err)
		}

		if len(admins) == 0 {
			return nil
		}

		adminModels := lo.Map(
			admins,
			func(admin Admin, _ int) *adminModel {
				return newAdminModel(model.ID, admin.UserID)
			},
		)

		if _, err := tx.NewInsert().
			Model(&adminModels).
			Exec(ctx); err != nil {
			return fmt.Errorf("failed to create or update group admins: %w", err)
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("failed to create or update group: %w", err)
	}

	return nil
}

// UpdateStatus implements Repository.
func (r *Repository) UpdateStatus(ctx context.Context, telegramID int64, isActive bool) error {
	_, err := r.db.NewUpdate().
		Model((*GroupModel)(nil)).
		Set("is_active = ?", isActive).
		Where("telegram_group_id = ?", telegramID).
		Exec(ctx)

	if err != nil {
		return fmt.Errorf("failed to update group status: %w", err)
	}

	return nil
}
