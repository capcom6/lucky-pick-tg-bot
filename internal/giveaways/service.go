package giveaways

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"time"

	"github.com/go-telegram/bot"
	"go.uber.org/zap"
)

type Service struct {
	giveaways *Repository

	bot *bot.Bot

	logger *zap.Logger
}

func NewService(giveaways *Repository, bot *bot.Bot, logger *zap.Logger) *Service {
	return &Service{
		giveaways: giveaways,

		bot: bot,

		logger: logger,
	}
}

func (s *Service) ListScheduled(ctx context.Context) ([]Giveaway, error) {
	items, err := s.giveaways.ListScheduled(ctx)
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
		if winErr != nil {
			newStatus = NewCancelledGiveaway(giveaway.ID)
		} else {
			newStatus = NewFinishedGiveaway(giveaway.ID, winner.UserID)
		}

		if updErr := s.giveaways.Update(
			ctx,
			newStatus,
		); updErr != nil {
			logger.Error("failed to update giveaway",
				zap.Error(updErr),
			)
		}

		winners = append(winners, Winner{
			Giveaway:    newGiveaway(giveaway),
			Participant: newParticipant(winner),
		})
	}

	return winners, nil
}

func (s *Service) Published(ctx context.Context, id, messageID int64) error {
	return s.giveaways.Update(
		ctx,
		NewPublishGiveaway(
			id,
			messageID,
		),
	)
}

func (s *Service) Pending(ctx context.Context, id int64) error {
	return s.giveaways.Update(
		ctx,
		NewClosedGiveaway(
			id,
		),
	)
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

	return s.giveaways.AddParticipant(ctx, NewParticipantModel(giveawayID, userID))
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
