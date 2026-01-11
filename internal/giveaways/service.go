package giveaways

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"time"

	"github.com/capcom6/lucky-pick-tg-bot/internal/actions"
	"github.com/capcom6/lucky-pick-tg-bot/internal/groups"
	"github.com/samber/lo"
	"go.uber.org/zap"
)

type Service struct {
	giveaways *Repository

	llmSvc     *LLM
	groupsSvc  *groups.Service
	actionsSvc *actions.Service

	logger *zap.Logger
}

func NewService(
	giveaways *Repository,
	llmSvc *LLM,
	groupsSvc *groups.Service,
	actionsSvc *actions.Service,
	logger *zap.Logger,
) *Service {
	return &Service{
		giveaways: giveaways,

		llmSvc:     llmSvc,
		groupsSvc:  groupsSvc,
		actionsSvc: actionsSvc,

		logger: logger,
	}
}

func (s *Service) GenerateDescription(
	ctx context.Context,
	description string,
	publishDate time.Time,
	photo []byte,
) (string, error) {
	description, err := s.llmSvc.MakeDescription(ctx, description, publishDate, photo)
	if err != nil {
		return "", fmt.Errorf("failed to generate description: %w", err)
	}

	return description, nil
}

func (s *Service) Create(ctx context.Context, giveaway GiveawayPrepared) error {
	return s.giveaways.Create(ctx, giveaway)
}

func (s *Service) ListByIDs(ctx context.Context, giveawayIDs []int64) ([]Giveaway, error) {
	items, err := s.giveaways.ListByIDs(ctx, giveawayIDs)
	if err != nil {
		return nil, err
	}

	grps, err := s.selectGroups(ctx, items)
	if err != nil {
		return nil, err
	}

	return mapGiveaways(items, grps)
}

func (s *Service) ListReadyToPublish(ctx context.Context) ([]Giveaway, error) {
	items, err := s.giveaways.ListReadyToPublish(ctx)
	if err != nil {
		return nil, err
	}

	grps, err := s.selectGroups(ctx, items)
	if err != nil {
		return nil, err
	}

	return mapGiveaways(items, grps)
}

func (s *Service) ListActive(ctx context.Context) ([]Giveaway, error) {
	items, err := s.giveaways.ListActive(ctx)
	if err != nil {
		return nil, err
	}

	grps, err := s.selectGroups(ctx, items)
	if err != nil {
		return nil, err
	}

	return mapGiveaways(items, grps)
}

func (s *Service) ListApplicationFinished(ctx context.Context) ([]Giveaway, error) {
	items, err := s.giveaways.ListApplicationFinished(ctx)
	if err != nil {
		return nil, err
	}

	grps, err := s.selectGroups(ctx, items)
	if err != nil {
		return nil, err
	}

	return mapGiveaways(items, grps)
}

func (s *Service) ListWinners(ctx context.Context) ([]Winner, error) {
	giveaways, err := s.giveaways.ListResultsWait(ctx)
	if err != nil {
		return nil, err
	}

	grps, err := s.selectGroups(ctx, giveaways)
	if err != nil {
		return nil, err
	}

	winners := make([]Winner, 0)
	for _, giveaway := range giveaways {
		logger := s.logger.With(zap.Int64("giveaway_id", giveaway.ID))

		g, ok := grps[giveaway.GroupID]
		if !ok {
			logger.Error("group not found", zap.Int64("group_id", giveaway.GroupID))
			continue
		}

		logger.Debug("starting winner selection")
		winner, winErr := s.randomWinner(ctx, &giveaway)
		if winErr != nil && !errors.Is(winErr, ErrNotEnoughParticipants) {
			logger.Error("failed to generate random winner",
				zap.Int("participants_count", len(giveaway.Participants)),
				zap.Error(winErr),
			)
			continue
		}

		var newStatus *GiveawayModel
		var actionType, actionDesc string
		if winErr != nil {
			newStatus = NewCancelGiveaway(giveaway.ID)
			actionType = "giveaway.cancelled"
			actionDesc = fmt.Sprintf("Cancel giveaway: %s", winErr.Error())
		} else {
			logger.Debug(
				"winner selected successfully",
				zap.Int64("giveaway_id", giveaway.ID),
				zap.Int64("winner_user_id", winner.UserID),
			)
			newStatus = NewFinishGiveaway(giveaway.ID, winner.UserID)
			actionType = "giveaway.finished"
			actionDesc = "Finish giveaway"
		}

		if updErr := s.giveaways.Update(
			ctx,
			newStatus,
		); updErr != nil {
			logger.Error("failed to update giveaway",
				zap.Error(updErr),
			)
			continue
		}
		logger.Debug(
			"giveaway updated successfully",
			zap.Int64("giveaway_id", giveaway.ID),
			zap.String("status", string(newStatus.Status)),
		)

		s.actionsSvc.LogAction(
			ctx,
			actionType,
			newStatus.WinnerUserID,
			giveaway.ID,
			actionDesc,
		)

		winners = append(winners, Winner{
			Giveaway:    *newGiveaway(giveaway, g),
			Participant: newParticipant(winner),
		})
	}

	return winners, nil
}

func (s *Service) Published(ctx context.Context, id, messageID int64) error {
	if err := s.giveaways.Update(
		ctx,
		NewPublishGiveaway(
			id,
			messageID,
		),
	); err != nil {
		return err
	}

	// Log the action
	s.actionsSvc.LogAction(
		ctx,
		"giveaway.published",
		0,
		id,
		fmt.Sprintf("Publish giveaway with message ID %d", messageID),
	)

	return nil
}

func (s *Service) Close(ctx context.Context, id int64) error {
	if err := s.giveaways.Update(
		ctx,
		NewCloseGiveaway(
			id,
		),
	); err != nil {
		return err
	}

	// Log the action
	s.actionsSvc.LogAction(ctx, "giveaway.closed", 0, id, "Close giveaway")

	return nil
}

func (s *Service) Participate(ctx context.Context, giveawayID int64, userID int64) error {
	giveaway, err := s.giveaways.GetByID(ctx, giveawayID)
	if err != nil {
		return err
	}

	now := time.Now()

	if giveaway.PublishDate.After(now) {
		return ErrNotFound
	}

	if giveaway.ApplicationEndDate.Before(now) {
		return ErrNotFound
	}

	if addErr := s.giveaways.AddParticipant(ctx, NewParticipantModel(giveawayID, userID)); addErr != nil {
		return addErr
	}

	// Log the action
	s.actionsSvc.LogAction(ctx, "giveaway.participated", userID, giveawayID, "Participate in giveaway")

	return nil
}

func (s *Service) randomWinner(_ context.Context, giveaway *GiveawayModel) (*ParticipantModel, error) {
	if len(giveaway.Participants) == 0 {
		return nil, ErrNotEnoughParticipants
	}

	if len(giveaway.Participants) == 1 {
		return giveaway.Participants[0], nil
	}

	idx, err := rand.Int(rand.Reader, big.NewInt(int64(len(giveaway.Participants))))
	if err != nil {
		return nil, fmt.Errorf("failed to generate random number: %w", err)
	}

	return giveaway.Participants[idx.Int64()], nil
}

func (s *Service) selectGroups(ctx context.Context, items []GiveawayModel) (map[int64]groups.GroupWithSettings, error) {
	if len(items) == 0 {
		return map[int64]groups.GroupWithSettings{}, nil
	}

	ids := lo.UniqMap(
		items,
		func(item GiveawayModel, _ int) int64 {
			return item.GroupID
		},
	)

	grps, err := s.groupsSvc.SelectByIDs(ctx, ids)
	if err != nil {
		return nil, fmt.Errorf("failed to list groups: %w", err)
	}

	return lo.KeyBy(grps, func(item groups.GroupWithSettings) int64 {
		return item.ID
	}), nil
}

func (s *Service) SettingsForGroup(ctx context.Context, id int64) (*Settings, error) {
	group, err := s.groupsSvc.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get group: %w", err)
	}

	settings, err := NewSettings(group.Settings)
	if err != nil {
		return nil, fmt.Errorf("failed to parse settings: %w", err)
	}

	return &settings, nil
}
