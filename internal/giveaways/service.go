package giveaways

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"time"

	"github.com/capcom6/lucky-pick-tg-bot/internal/actions"
	"github.com/go-telegram/bot"
	"go.uber.org/zap"
)

type Service struct {
	giveaways *Repository

	bot        *bot.Bot
	actionsSvc *actions.Service

	logger *zap.Logger
}

func NewService(giveaways *Repository, bot *bot.Bot, logger *zap.Logger, actionsSvc *actions.Service) *Service {
	return &Service{
		giveaways: giveaways,

		bot:        bot,
		actionsSvc: actionsSvc,

		logger: logger,
	}
}

func (s *Service) ListByIDs(ctx context.Context, giveawayIDs []int64) ([]Giveaway, error) {
	items, err := s.giveaways.ListByIDs(ctx, giveawayIDs)
	if err != nil {
		return nil, err
	}

	return mapGiveaways(items), nil
}

func (s *Service) ListReadyToPublish(ctx context.Context) ([]Giveaway, error) {
	items, err := s.giveaways.ListReadyToPublish(ctx)
	if err != nil {
		return nil, err
	}

	return mapGiveaways(items), nil
}

func (s *Service) ListActive(ctx context.Context) ([]Giveaway, error) {
	items, err := s.giveaways.ListActive(ctx)
	if err != nil {
		return nil, err
	}

	return mapGiveaways(items), nil
}

func (s *Service) ListApplicationFinished(ctx context.Context) ([]Giveaway, error) {
	items, err := s.giveaways.ListApplicationFinished(ctx)
	if err != nil {
		return nil, err
	}

	return mapGiveaways(items), nil
}

func (s *Service) ListWinners(ctx context.Context) ([]Winner, error) {
	giveaways, err := s.giveaways.ListResultsWait(ctx)
	if err != nil {
		return nil, err
	}

	winners := make([]Winner, 0)
	for _, giveaway := range giveaways {
		logger := s.logger.With(zap.Int64("giveaway_id", giveaway.ID))
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

		s.actionsSvc.LogAction(
			ctx,
			actionType,
			newStatus.WinnerUserID,
			giveaway.ID,
			actionDesc,
		)

		winners = append(winners, Winner{
			Giveaway:    *newGiveaway(giveaway),
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

func (s *Service) Create(ctx context.Context, giveaway GiveawayDraft) error {
	model := NewGiveawayModel(
		giveaway.GroupID, giveaway.AdminUserID,
		giveaway.PhotoFileID, giveaway.Description,
		giveaway.PublishDate, giveaway.ApplicationEndDate, giveaway.ResultsDate,
		giveaway.IsAnonymous,
	)

	return s.giveaways.Create(ctx, model)
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
