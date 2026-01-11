package handlers

import (
	"strconv"
	"strings"

	"github.com/capcom6/lucky-pick-tg-bot/internal/bot/adaptor"
	"github.com/capcom6/lucky-pick-tg-bot/internal/bot/handler"
	"github.com/capcom6/lucky-pick-tg-bot/internal/giveaways"
	"github.com/capcom6/lucky-pick-tg-bot/internal/users"
	"github.com/capcom6/lucky-pick-tg-bot/pkg/gotelegrambotfx"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"go.uber.org/zap"
)

const participatePrefix = "participate:"

type Participant struct {
	handler.BaseHandler

	usersSvc     *users.Service
	giveawaysSvc *giveaways.Service
}

func NewParticipant(
	bot *gotelegrambotfx.Bot,
	usersSvc *users.Service,
	giveawaysSvc *giveaways.Service,
	logger *zap.Logger,
) handler.Handler {
	return &Participant{
		BaseHandler: handler.BaseHandler{
			Bot:    bot,
			Logger: logger,
		},

		usersSvc:     usersSvc,
		giveawaysSvc: giveawaysSvc,
	}
}

func (p *Participant) Register(b *gotelegrambotfx.Bot) {
	b.RegisterHandlerMatchFunc(p.filterParticipateCallback, adaptor.New(p.handleParticipate))
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

func (p *Participant) handleParticipate(ctx *adaptor.Context, update *models.Update) {
	if update.CallbackQuery == nil {
		p.Logger.Error("invalid update: missing callback query")
		return
	}

	logger := p.WithContext(update)

	alertText := "Ваша заявка принята!"

	defer func() {
		if _, err := p.Bot.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{
			CallbackQueryID: update.CallbackQuery.ID,
			Text:            alertText,
			ShowAlert:       true,
		}); err != nil {
			logger.Error("failed to answer callback query", zap.Error(err))
		}
	}()

	user, err := ctx.User()
	if err != nil {
		alertText = alertSomethingWrong
		logger.Error("failed to get user", zap.Error(err))
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
