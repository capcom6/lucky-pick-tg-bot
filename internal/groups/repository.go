package groups

import (
	"context"
	"database/sql"
	"errors"
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
func (r *Repository) CreateOrUpdate(ctx context.Context, group *GroupDraft, admins []Admin) error {
	err := r.db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
		model := newGroupModel(group)

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

func (r *Repository) SelectByIDs(ctx context.Context, ids []int64) ([]GroupWithSettings, error) {
	if len(ids) == 0 {
		return []GroupWithSettings{}, nil
	}

	var groups []GroupModel
	if err := r.db.NewSelect().
		Model(&groups).
		Distinct().
		Where("id IN (?)", bun.In(ids)).
		Relation("Settings").
		Scan(ctx); err != nil {
		return nil, fmt.Errorf("failed to get groups by IDs: %w", err)
	}

	return lo.Map(
		groups,
		func(m GroupModel, _ int) GroupWithSettings {
			return *newGroupWithSettings(&m)
		}), nil
}

// GetByUser returns groups where the user is an admin.
func (r *Repository) GetByUser(ctx context.Context, userID int64) ([]GroupWithSettings, error) {
	var groups []GroupModel
	if err := r.db.NewSelect().
		Model(&groups).
		Join("JOIN group_admins ga ON ga.group_id = g.id").
		Where("g.is_active = ?", true).
		Where("ga.user_id = ?", userID).
		Relation("Settings").
		Scan(ctx); err != nil {
		return nil, fmt.Errorf("failed to get user admin groups: %w", err)
	}

	return lo.Map(
		groups,
		func(m GroupModel, _ int) GroupWithSettings {
			return *newGroupWithSettings(&m)
		}), nil
}

func (r *Repository) IsAdmin(ctx context.Context, groupID int64, userID int64) (bool, error) {
	var admin adminModel
	err := r.db.NewSelect().
		Model(&admin).
		Where("group_id = ?", groupID).
		Where("user_id = ?", userID).
		Scan(ctx)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}

	if err != nil {
		return false, fmt.Errorf("failed to check if user is admin: %w", err)
	}

	return true, nil
}

// GetByID returns a group by its ID.
func (r *Repository) GetByID(ctx context.Context, groupID int64) (*GroupWithSettings, error) {
	group := new(GroupModel)
	err := r.db.NewSelect().
		Model(group).
		Relation("Settings").
		Where("id = ?", groupID).
		Scan(ctx)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get group by ID: %w", err)
	}

	return newGroupWithSettings(group), nil
}

// UpdateStatus updates the status of a group.
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

// UpdateSettings updates multiple settings of a group at once.
func (r *Repository) UpdateSettings(ctx context.Context, groupID int64, settings map[string]string) error {
	rows := lo.MapToSlice(
		settings,
		func(key, value string) *settingsModel {
			return newSettingsModel(groupID, key, value)
		},
	)

	_, err := r.db.NewInsert().
		Model(&rows).
		On("DUPLICATE KEY UPDATE").
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to update settings: %w", err)
	}

	return nil
}

// GetAllSettings returns all settings for a group as a map.
func (r *Repository) GetAllSettings(ctx context.Context, groupID int64) (map[string]string, error) {
	var settings []settingsModel
	err := r.db.NewSelect().
		Model(&settings).
		Where("group_id = ?", groupID).
		Scan(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get settings for group %d: %w", groupID, err)
	}

	result := make(map[string]string, len(settings))
	for _, setting := range settings {
		result[setting.Key] = setting.Value
	}

	return result, nil
}

// DeleteSetting removes a setting for a group.
func (r *Repository) DeleteSetting(ctx context.Context, groupID int64, key string) error {
	_, err := r.db.NewDelete().
		Model((*settingsModel)(nil)).
		Where("group_id = ?", groupID).
		Where("`key` = ?", key).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to delete setting %s: %w", key, err)
	}

	return nil
}
