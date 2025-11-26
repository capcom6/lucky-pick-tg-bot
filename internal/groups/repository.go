package groups

import (
	"context"
	"fmt"

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
func (r *Repository) CreateOrUpdate(ctx context.Context, group *Group) error {
	model := NewGroupModel(group.TelegramID, group.Title)

	_, err := r.db.NewInsert().
		Model(model).
		On("DUPLICATE KEY UPDATE").
		Returning("*").
		Exec(ctx)

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
