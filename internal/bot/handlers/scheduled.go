package handlers

import (
	"fmt"
	"strings"

	"github.com/capcom6/lucky-pick-tg-bot/internal/bot/adaptor"
	"github.com/capcom6/lucky-pick-tg-bot/internal/bot/handler"
	"github.com/capcom6/lucky-pick-tg-bot/internal/giveaways"
	"github.com/capcom6/lucky-pick-tg-bot/internal/groups"
	"github.com/capcom6/lucky-pick-tg-bot/pkg/gotelegrambotfx"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/samber/lo"
	"go.uber.org/zap"
)

const scheduledCommand = "/scheduled"

type Scheduled struct {
	handler.BaseHandler

	groupsSvc    *groups.Service
	giveawaysSvc *giveaways.Service
}

func NewScheduled(
	bot *gotelegrambotfx.Bot,
	groupsSvc *groups.Service,
	giveawaysSvc *giveaways.Service,
	logger *zap.Logger,
) handler.Handler {
	return &Scheduled{
		BaseHandler: handler.BaseHandler{
			Bot:    bot,
			Logger: logger,
		},
		groupsSvc:    groupsSvc,
		giveawaysSvc: giveawaysSvc,
	}
}

func (s *Scheduled) Register(b *gotelegrambotfx.Bot) {
	b.RegisterHandlerMatchFunc(
		s.filterScheduledCommand,
		adaptor.New(s.handleScheduledCommand),
	)
}

func (s *Scheduled) filterScheduledCommand(update *models.Update) bool {
	if update.Message == nil || update.Message.Text == "" {
		return false
	}

	return update.Message.Text == scheduledCommand && update.Message.Chat.Type == models.ChatTypePrivate
}

func (s *Scheduled) handleScheduledCommand(ctx *adaptor.Context, update *models.Update) {
	logger := s.WithContext(update)

	user, err := ctx.User()
	if err != nil {
		logger.Error("failed to get user", zap.Error(err))
		s.SendReply(ctx, update, &bot.SendMessageParams{Text: "❌ Failed to process user. Please try again."})
		return
	}

	adminGroups, err := s.groupsSvc.GetUserAdminGroups(ctx, user.ID)
	if err != nil {
		logger.Error("failed to get user admin groups", zap.Error(err))
		s.SendReply(ctx, update, &bot.SendMessageParams{Text: "❌ Failed to verify admin status. Please try again."})
		return
	}

	if len(adminGroups) == 0 {
		s.SendReply(
			ctx,
			update,
			&bot.SendMessageParams{Text: "❌ You must be an admin of a group to view scheduled giveaways."},
		)
		return
	}

	groupIDs := lo.Map(adminGroups, func(group groups.GroupWithSettings, _ int) int64 {
		return group.ID
	})

	items, err := s.giveawaysSvc.ListScheduledByGroupIDs(ctx, groupIDs)
	if err != nil {
		logger.Error("failed to list scheduled giveaways", zap.Error(err))
		s.SendReply(
			ctx,
			update,
			&bot.SendMessageParams{Text: "❌ Failed to load scheduled giveaways. Please try again."},
		)
		return
	}

	if len(items) == 0 {
		s.SendReply(ctx, update, &bot.SendMessageParams{Text: "📭 No scheduled giveaways found."})
		return
	}

	const descLen = 100
	var text strings.Builder
	text.WriteString("🗓 *Scheduled giveaways*\n\n")

	for i, item := range items {
		fmt.Fprintf(&text, "%d\\. *%s*\nDescription: `%s`\nStart: `%s`\nApplications end: `%s`\nResults: `%s`\n\n",
			i+1,
			bot.EscapeMarkdown(item.Group.Title),
			bot.EscapeMarkdown(lo.Substring(item.Description, 0, descLen)),
			bot.EscapeMarkdown(formatDateTime(item.PublishDate)),
			bot.EscapeMarkdown(formatDateTime(item.ApplicationEndDate)),
			bot.EscapeMarkdown(formatDateTime(item.ResultsDate)))
	}

	s.SendReply(
		ctx,
		update,
		&bot.SendMessageParams{
			Text:      text.String(),
			ParseMode: models.ParseModeMarkdown,
		},
	)
}
