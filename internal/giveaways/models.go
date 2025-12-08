package giveaways

import (
	"time"

	"github.com/capcom6/lucky-pick-tg-bot/internal/groups"
	"github.com/capcom6/lucky-pick-tg-bot/internal/users"
	"github.com/uptrace/bun"
)

type Status string

const (
	StatusScheduled Status = "scheduled"
	StatusActive    Status = "active"
	StatusClosed    Status = "closed"
	StatusFinished  Status = "finished"
	StatusCancelled Status = "cancelled"
)

type GiveawayModel struct {
	bun.BaseModel `bun:"table:giveaways,alias:ga"`

	ID int64 `bun:"id,pk,autoincrement"`

	GroupID            int64     `bun:"group_id,notnull"`
	AdminUserID        int64     `bun:"admin_user_id,notnull"`
	PhotoFileID        string    `bun:"photo_file_id,notnull"`
	Description        string    `bun:"description,notnull"`
	PublishDate        time.Time `bun:"publish_date,notnull"`
	ApplicationEndDate time.Time `bun:"application_end_date,notnull"`
	ResultsDate        time.Time `bun:"results_date,notnull"`
	IsAnonymous        bool      `bun:"is_anonymous,notnull"`

	TelegramMessageID int64 `bun:"telegram_message_id,nullzero"`

	WinnerUserID int64  `bun:"winner_user_id,nullzero"`
	Status       Status `bun:"status,notnull,default:'scheduled'"`

	CreatedAt time.Time `bun:"created_at,scanonly"`
	UpdatedAt time.Time `bun:"updated_at,scanonly"`

	Group        *groups.GroupModel  `bun:"g,rel:belongs-to,join:group_id=id"`
	Participants []*ParticipantModel `bun:"gap,rel:has-many,join:id=giveaway_id"`
}

func NewGiveawayModel(
	groupID, adminUserID int64,
	photoFileID, description string,
	publishDate, applicationEndDate, resultsDate time.Time,
	isAnonymous bool,
) *GiveawayModel {
	//nolint:exhaustruct // partial constructor
	return &GiveawayModel{
		GroupID:     groupID,
		AdminUserID: adminUserID,

		PhotoFileID:        photoFileID,
		Description:        description,
		PublishDate:        publishDate,
		ApplicationEndDate: applicationEndDate,
		ResultsDate:        resultsDate,

		IsAnonymous: isAnonymous,
	}
}

func NewPublishGiveaway(id, messageID int64) *GiveawayModel {
	//nolint:exhaustruct // partial constructor
	return &GiveawayModel{
		ID:                id,
		TelegramMessageID: messageID,
		Status:            StatusActive,
	}
}

func NewCloseGiveaway(id int64) *GiveawayModel {
	//nolint:exhaustruct // partial constructor
	return &GiveawayModel{
		ID:     id,
		Status: StatusClosed,
	}
}

func NewFinishGiveaway(id, winnerID int64) *GiveawayModel {
	//nolint:exhaustruct // partial constructor
	return &GiveawayModel{
		ID:           id,
		WinnerUserID: winnerID,
		Status:       StatusFinished,
	}
}

func NewCancelGiveaway(id int64) *GiveawayModel {
	//nolint:exhaustruct // partial constructor
	return &GiveawayModel{
		ID:     id,
		Status: StatusCancelled,
	}
}

type ParticipantModel struct {
	bun.BaseModel `bun:"table:giveaway_participants,alias:gap"`

	ID         int64 `bun:"id,pk,autoincrement"`
	GiveawayID int64 `bun:"giveaway_id,notnull"`
	UserID     int64 `bun:"user_id,notnull"`

	JoinedAt time.Time `bun:"joined_at,scanonly"`

	User *users.UserModel `bun:"u,rel:belongs-to,join:user_id=id"`
}

func NewParticipantModel(giveawayID, userID int64) *ParticipantModel {
	//nolint:exhaustruct // partial constructor
	return &ParticipantModel{
		GiveawayID: giveawayID,
		UserID:     userID,
	}
}
