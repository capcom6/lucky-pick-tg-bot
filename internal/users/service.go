package users

import (
	"context"
	"fmt"

	"github.com/capcom6/lucky-pick-tg-bot/internal/actions"
	"go.uber.org/zap"
)

type Service struct {
	users *Repository

	actionsSvc *actions.Service

	logger *zap.Logger
}

func NewService(users *Repository, logger *zap.Logger, actionsSvc *actions.Service) *Service {
	return &Service{
		users: users,

		actionsSvc: actionsSvc,

		logger: logger,
	}
}

// RegisterUser creates or updates a user from Telegram user data.
func (s *Service) RegisterUser(ctx context.Context, user UserIn) (*User, error) {
	logger := s.logger.With(
		zap.Int64("telegram_user_id", user.TelegramUserID),
	)

	model := NewUserModel(
		user.TelegramUserID,
		user.Username,
		user.FirstName,
		user.LastName,
	)

	// Create or update user in database
	created, err := s.users.CreateOrUpdate(ctx, model)
	if err != nil {
		logger.Error("failed to create or update user",
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to register user: %w", err)
	}

	logger = logger.With(
		zap.Int64("user_id", model.ID),
	)

	if created {
		logger.Info("user registered successfully",
			zap.String("username", model.Username),
			zap.String("first_name", model.FirstName),
		)

		// Log action after successful DB operation
		s.actionsSvc.LogAction(ctx, "user.registered", model.ID, 0, fmt.Sprintf("Registered user @%s", model.Username))
	}

	return &User{
		UserIn: UserIn{
			TelegramUserID: model.TelegramUserID,
			Username:       model.Username,
			FirstName:      model.FirstName,
			LastName:       model.LastName,
		},
		ID:           model.ID,
		RegisteredAt: model.RegisteredAt,
		IsActive:     model.IsActive,
	}, nil
}
