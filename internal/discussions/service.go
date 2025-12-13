package discussions

import (
	"context"
	"fmt"
	"time"

	"github.com/capcom6/lucky-pick-tg-bot/internal/giveaways"
	"github.com/capcom6/lucky-pick-tg-bot/internal/llm"
	"github.com/samber/lo"
	"go.uber.org/zap"
)

type Service struct {
	discussions *Repository

	giveawaysSvc *giveaways.Service
	llmSvc       *llm.Service

	logger *zap.Logger
}

func NewService(
	discussions *Repository,
	giveawaysSvc *giveaways.Service,
	llmSvc *llm.Service,
	logger *zap.Logger,
) *Service {
	return &Service{
		discussions: discussions,

		giveawaysSvc: giveawaysSvc,
		llmSvc:       llmSvc,

		logger: logger,
	}
}

func (s *Service) Generate(ctx context.Context) ([]Discussion, error) {
	givs, err := s.giveawaysSvc.ListActive(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list giveaways: %w", err)
	}

	givs = lo.Filter(
		givs,
		func(item giveaways.Giveaway, _ int) bool {
			return item.Group.DiscussionsThreshold > 0 &&
				time.Since(item.PublishDate) > time.Hour*time.Duration(item.Group.DiscussionsThreshold)
		},
	)

	if len(givs) == 0 {
		return []Discussion{}, nil
	}

	indexed := lo.KeyBy(
		givs,
		func(item giveaways.Giveaway) int64 {
			return item.ID
		},
	)
	discuss, err := s.discussions.ListByGiveaway(ctx, lo.Keys(indexed))
	if err != nil {
		return nil, fmt.Errorf("failed to list discussions: %w", err)
	}

	for _, d := range discuss {
		delete(indexed, d.GiveawayID)
	}

	if len(indexed) == 0 {
		return []Discussion{}, nil
	}

	now := time.Now()
	questions := make([]Discussion, 0, len(indexed))
	for _, ga := range indexed {
		question, llmErr := s.llmSvc.MakeQuestion(ctx, ga.Description, now.Sub(ga.PublishDate))
		if llmErr != nil {
			s.logger.Error("failed to make question",
				zap.Int64("giveaway_id", ga.ID),
				zap.Error(llmErr),
			)
			continue
		}

		draft := DiscussionDraft{
			GiveawayID: ga.ID,
			UserID:     BotUserID,
			Text:       question,
		}

		if d, createErr := s.discussions.Create(ctx, draft); createErr != nil {
			s.logger.Error("failed to create discussion",
				zap.Int64("giveaway_id", ga.ID),
				zap.Error(createErr),
			)
		} else {
			questions = append(questions, *d)
		}
	}

	return questions, nil
}

func (s *Service) SetTelegramID(ctx context.Context, id int64, telegramID int64) error {
	return s.discussions.SetTelegramID(ctx, id, telegramID)
}
