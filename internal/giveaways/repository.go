package giveaways

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

func (r *Repository) ListScheduled(ctx context.Context) ([]GiveawayModel, error) {
	giveaways := make([]GiveawayModel, 0)
	if err := r.db.NewSelect().
		Model(&giveaways).
		Relation("Group").
		Where("ga.status = ?", StatusScheduled).
		Where("ga.publish_date <= NOW()").
		Where("g.is_active = ?", true).
		Scan(ctx); err != nil {
		return nil, fmt.Errorf("failed to get scheduled giveaways: %w", err)
	}

	return giveaways, nil
}

func (r *Repository) ListApplicationFinished(ctx context.Context) ([]GiveawayModel, error) {
	giveaways := make([]GiveawayModel, 0)
	if err := r.db.NewSelect().
		Model(&giveaways).
		Relation("Group").
		Where("ga.status = ?", StatusActive).
		Where("ga.application_end_date <= NOW()").
		Where("g.is_active = ?", true).
		Scan(ctx); err != nil {
		return nil, fmt.Errorf("failed to get application finished giveaways: %w", err)
	}

	return giveaways, nil
}

func (r *Repository) ListResultsWait(ctx context.Context) ([]GiveawayModel, error) {
	giveaways := make([]GiveawayModel, 0)
	if err := r.db.NewSelect().
		Model(&giveaways).
		Relation("Group").
		Relation("Participants").
		Relation("Participants.User").
		Where("ga.status = ?", StatusClosed).
		Where("ga.results_date <= NOW()").
		Where("g.is_active = ?", true).
		Scan(ctx); err != nil {
		return nil, fmt.Errorf("failed to get pending giveaways: %w", err)
	}

	return giveaways, nil
}

func (r *Repository) GetByID(ctx context.Context, giveawayID int64) (*GiveawayModel, error) {
	giveaway := new(GiveawayModel)
	if err := r.db.NewSelect().
		Model(giveaway).
		Relation("Group").
		Where("ga.id = ?", giveawayID).
		Scan(ctx); err != nil {
		return nil, fmt.Errorf("failed to get giveaway by ID: %w", err)
	}

	return giveaway, nil
}

func (r *Repository) Update(ctx context.Context, giveaway *GiveawayModel) error {
	_, err := r.db.NewUpdate().
		Model(giveaway).
		OmitZero().
		WherePK().
		Exec(ctx)

	if err != nil {
		return fmt.Errorf("failed to update giveaway: %w", err)
	}

	return nil
}

func (r *Repository) AddParticipant(ctx context.Context, participant *ParticipantModel) error {
	_, err := r.db.NewInsert().
		Ignore().
		Model(participant).
		Returning("*").
		Exec(ctx)

	if err != nil {
		return fmt.Errorf("failed to add participant: %w", err)
	}

	return nil
}
