package cancel

import (
	"github.com/capcom6/lucky-pick-tg-bot/internal/bot/adaptor"
	"github.com/capcom6/lucky-pick-tg-bot/internal/bot/handler"
	"github.com/capcom6/lucky-pick-tg-bot/pkg/gotelegrambotfx"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"go.uber.org/zap"
)

type Handler struct {
	handler.BaseHandler
}

func NewHandler(bot *gotelegrambotfx.Bot, logger *zap.Logger) handler.Handler {
	return &Handler{
		BaseHandler: handler.BaseHandler{
			Bot:    bot,
			Logger: logger,
		},
	}
}

// Register implements handler.Handler.
func (h *Handler) Register(b *gotelegrambotfx.Bot) {
	b.RegisterHandler(
		bot.HandlerTypeMessageText,
		"cancel",
		bot.MatchTypeCommandStartOnly,
		adaptor.New(h.handleCancel),
	)
}

func (h *Handler) handleCancel(ctx *adaptor.Context, update *models.Update) {
	state, err := ctx.State()
	if err != nil {
		h.Logger.Error("failed to get state", zap.Error(err))
		h.SendReply(ctx, update, &bot.SendMessageParams{
			Text: "❌ Failed to cancel operation. Please try again.",
		})
		return
	}

	state.Clear()

	h.SendReply(ctx, update, &bot.SendMessageParams{
		Text: "🔄 Operation cancelled.",
	})
}
