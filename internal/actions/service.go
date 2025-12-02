package actions

import (
	"context"

	"go.uber.org/zap"
)

type Service struct {
	actions *Repository

	logger *zap.Logger
}

func NewService(actions *Repository, logger *zap.Logger) *Service {
	return &Service{
		actions: actions,

		logger: logger,
	}
}

func (s *Service) LogAction(
	ctx context.Context,
	actionType string,
	userID int64,
	giveawayID int64,
	description string,
) {
	giveawayIDptr := ptrOrNil(giveawayID)
	userIDptr := ptrOrNil(userID)

	entry := NewEntry(
		giveawayIDptr,
		userIDptr,
		actionType,
		description,
	)

	if err := s.actions.LogAction(ctx, entry); err != nil {
		s.logger.Error("failed to log action", zap.Any("entry", entry), zap.Error(err))
		return
	}

	s.logger.Debug("action logged",
		zap.String("action_type", actionType),
		zap.String("description", description),
	)
}

func ptrOrNil(i int64) *int64 {
	if i == 0 {
		return nil
	}
	return &i
}
