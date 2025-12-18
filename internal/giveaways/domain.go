package giveaways

import (
	"fmt"
	"time"

	"github.com/capcom6/lucky-pick-tg-bot/internal/groups"
)

type GiveawayBase struct {
}

type GiveawayDraft struct {
	GroupID            int64
	AdminUserID        int64
	PhotoFileID        string
	Description        string
	PublishDate        time.Time
	ApplicationEndDate time.Time
	ResultsDate        time.Time
	IsAnonymous        bool
}

type GiveawayPrepared struct {
	GiveawayDraft

	OriginalDescription string
}

type Giveaway struct {
	GiveawayDraft

	ID int64

	Group groups.GroupWithSettings

	TelegramMessageID int64

	WinnerUserID int64
	Status       Status

	CreatedAt time.Time
	UpdatedAt time.Time
}

type Participant struct {
	ID int64

	UserID         int64
	UserTelegramID int64
	UserUsername   string
	UserFirstName  string

	JoinedAt time.Time
}

type Winner struct {
	Giveaway    Giveaway
	Participant *Participant
}

func newGiveaway(item GiveawayModel, group groups.GroupWithSettings) *Giveaway {
	return &Giveaway{
		GiveawayDraft: GiveawayDraft{
			GroupID:            item.GroupID,
			AdminUserID:        item.AdminUserID,
			PhotoFileID:        item.PhotoFileID,
			Description:        item.Description,
			PublishDate:        item.PublishDate,
			ApplicationEndDate: item.ApplicationEndDate,
			ResultsDate:        item.ResultsDate,
			IsAnonymous:        item.IsAnonymous,
		},

		ID: item.ID,

		Group: group,

		TelegramMessageID: item.TelegramMessageID,

		WinnerUserID: item.WinnerUserID,
		Status:       item.Status,

		CreatedAt: item.CreatedAt,
		UpdatedAt: item.UpdatedAt,
	}
}

func mapGiveaways(items []GiveawayModel, grps map[int64]groups.GroupWithSettings) ([]Giveaway, error) {
	result := make([]Giveaway, 0, len(items))

	for _, item := range items {
		g, ok := grps[item.GroupID]
		if !ok {
			return nil, fmt.Errorf("%w: ID %d", groups.ErrNotFound, item.GroupID)
		}

		result = append(result, *newGiveaway(item, g))
	}

	return result, nil
}

func newParticipant(item *ParticipantModel) *Participant {
	if item == nil {
		return nil
	}

	return &Participant{
		ID: item.ID,

		UserID:         item.UserID,
		UserTelegramID: item.User.TelegramUserID,
		UserUsername:   item.User.Username,
		UserFirstName:  item.User.FirstName,

		JoinedAt: item.JoinedAt,
	}
}
