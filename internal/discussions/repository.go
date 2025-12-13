package discussions

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/samber/lo"
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

func (r *Repository) ListByGiveaway(ctx context.Context, giveawayIDs []int64) ([]Discussion, error) {
	if len(giveawayIDs) == 0 {
		return []Discussion{}, nil
	}

	discussions := make([]discussionModel, 0)
	if err := r.db.NewSelect().
		Model(&discussions).
		Where("giveaway_id IN (?)", bun.In(giveawayIDs)).
		Scan(ctx); err != nil {
		return nil, fmt.Errorf("failed to get discussions: %w", err)
	}

	return lo.Map(discussions, func(discussion discussionModel, _ int) Discussion {
		return *discussion.toDiscussion()
	}), nil
}

func (r *Repository) ListLast(ctx context.Context, giveawayIDs []int64, after time.Time) ([]Discussion, error) {
	if len(giveawayIDs) == 0 {
		return []Discussion{}, nil
	}

	discussions := make([]discussionModel, 0)
	subquery := r.db.NewSelect().
		Model((*discussionModel)(nil)).
		ColumnExpr("giveaway_id, MAX(created_at) as created_at").
		Where("giveaway_id IN (?)", bun.In(giveawayIDs)).
		Where("created_at > ?", after).
		Group("giveaway_id")

	if err := r.db.NewSelect().
		Model(&discussions).
		TableExpr("(?) as last", subquery).
		Where("gd.created_at = last.created_at AND gd.giveaway_id = last.giveaway_id").
		Scan(ctx); err != nil {
		return nil, fmt.Errorf("failed to get last discussions: %w", err)
	}

	return lo.Map(discussions, func(discussion discussionModel, _ int) Discussion {
		return *discussion.toDiscussion()
	}), nil
}

func (r *Repository) GetLast(ctx context.Context, giveawayID int64, after time.Time) (*Discussion, error) {
	discussion := new(discussionModel)
	if err := r.db.NewSelect().
		Model(discussion).
		Where("giveaway_id = ?", giveawayID).
		Where("created_at > ?", after).
		Order("created_at DESC").
		Limit(1).
		Scan(ctx); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("failed to get last discussion: %w", err)
	}

	return discussion.toDiscussion(), nil
}

func (r *Repository) Create(ctx context.Context, discussion DiscussionDraft) (*Discussion, error) {
	model := newDiscussionModel(discussion)

	_, err := r.db.NewInsert().
		Model(model).
		Returning("*").
		Exec(ctx)

	if err != nil {
		return nil, fmt.Errorf("failed to create discussion: %w", err)
	}

	return model.toDiscussion(), nil
}

func (r *Repository) SetTelegramID(ctx context.Context, id int64, telegramID int64) error {
	_, err := r.db.NewUpdate().
		Model((*discussionModel)(nil)).
		Set("telegram_message_id = ?", telegramID).
		Where("id = ?", id).
		Exec(ctx)

	if err != nil {
		return fmt.Errorf("failed to set telegram ID: %w", err)
	}

	return nil
}
