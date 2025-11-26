package handlers

import (
	"context"
	"strconv"
	"strings"

	"github.com/capcom6/lucky-pick-tg-bot/internal/giveaways"
	"github.com/capcom6/lucky-pick-tg-bot/internal/users"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"go.uber.org/zap"
)

const participatePrefix = "participate:"

type Participant struct {
	BaseHandler

	usersSvc     *users.Service
	giveawaysSvc *giveaways.Service
}

func NewParticipant(
	bot *bot.Bot,
	usersSvc *users.Service,
	giveawaysSvc *giveaways.Service,
	logger *zap.Logger,
) Handler {
	return &Participant{
		BaseHandler: BaseHandler{
			bot:    bot,
			logger: logger,
		},

		usersSvc:     usersSvc,
		giveawaysSvc: giveawaysSvc,
	}
}

func (p *Participant) Register(b *bot.Bot) {
	b.RegisterHandlerMatchFunc(p.filterParticipateCallback, p.handleParticipate)
}

func (p *Participant) filterParticipateCallback(update *models.Update) bool {
	if update.CallbackQuery == nil {
		return false
	}

	if !strings.HasPrefix(update.CallbackQuery.Data, participatePrefix) {
		return false
	}

	return true
}

func (p *Participant) handleParticipate(ctx context.Context, b *bot.Bot, update *models.Update) {
	if update.CallbackQuery == nil {
		p.logger.Error("invalid update: missing callback query")
		return
	}

	logger := p.withContext(update)

	alertText := "Ваша заявка принята!"

	defer func() {
		if _, err := b.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{
			CallbackQueryID: update.CallbackQuery.ID,
			Text:            alertText,
			ShowAlert:       true,
		}); err != nil {
			logger.Error("failed to answer callback query", zap.Error(err))
		}
	}()

	user, err := p.usersSvc.RegisterUser(ctx, UserToDomain(&update.CallbackQuery.From))
	if err != nil {
		alertText = alertSomethingWrong
		logger.Error("failed to register user", zap.Error(err))
		return
	}

	logger = logger.With(zap.Int64("user_id", user.ID))

	// Extract giveaway ID from callback data
	giveawayIDStr := strings.TrimPrefix(update.CallbackQuery.Data, participatePrefix)
	giveawayID, err := strconv.ParseInt(giveawayIDStr, 10, 64)
	if err != nil {
		alertText = alertSomethingWrong
		logger.Error("failed to parse giveaway ID", zap.Error(err))
		return
	}

	logger = logger.With(zap.Int64("giveaway_id", giveawayID))

	if participateErr := p.giveawaysSvc.Participate(ctx, giveawayID, user.ID); participateErr != nil {
		alertText = alertSomethingWrong
		logger.Error("failed to participate in giveaway", zap.Error(participateErr))
		return
	}
}
