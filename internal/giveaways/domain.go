package giveaways

import (
	"time"

	"github.com/samber/lo"
)

type Giveaway struct {
	ID int64

	GroupID         int64
	TelegramGroupID int64

	AdminUserID        int64
	PhotoFileID        string
	Description        string
	PublishDate        time.Time
	ApplicationEndDate time.Time
	ResultsDate        time.Time
	IsAnonymous        bool

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

func newGiveaway(item GiveawayModel) Giveaway {
	return Giveaway{
		ID: item.ID,

		GroupID:         item.GroupID,
		TelegramGroupID: item.Group.TelegramGroupID,

		AdminUserID:        item.AdminUserID,
		PhotoFileID:        item.PhotoFileID,
		Description:        item.Description,
		PublishDate:        item.PublishDate,
		ApplicationEndDate: item.ApplicationEndDate,
		ResultsDate:        item.ResultsDate,
		IsAnonymous:        item.IsAnonymous,

		TelegramMessageID: item.TelegramMessageID,

		WinnerUserID: item.WinnerUserID,
		Status:       item.Status,

		CreatedAt: item.CreatedAt,
		UpdatedAt: item.UpdatedAt,
	}
}

func mapGiveaways(items []GiveawayModel) []Giveaway {
	return lo.Map(
		items,
		func(item GiveawayModel, _ int) Giveaway {
			return newGiveaway(item)
		},
	)
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
