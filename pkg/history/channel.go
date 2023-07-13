package history

import (
	"context"
	"time"

	"github.com/pkg/errors"
	"github.com/vennekilde/gw2verify/internal/api"
	"github.com/vennekilde/gw2verify/internal/orm"
	"github.com/vennekilde/gw2verify/pkg/verify"
)

type ChannelStatistics struct {
	Timestamp     time.Time
	Users         int
	MutedUsers    int
	DeafenedUsers int
	UnrankedUsers int
	RankedUsers   int
}

type VoiceUserState struct {
	Timestamp          time.Time
	ServiceID          int
	ServiceUserID      string
	ChannelID          string
	Muted              bool
	Deafened           bool
	WvWRank            int `json:"wvw_rank" bun:"wvw_rank"`
	Age                int
	VerificationStatus int
}

func CollectChannelStatistics(serviceID int, channelID string, worldPerspective int, data api.ChannelMetadata) error {
	ctx := context.Background()
	db := orm.DB()

	// Begin transaction
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	committed := false
	defer func() {
		if !committed {
			tx.Rollback()
		}
	}()

	ts := time.Now()
	for _, userMetadata := range data.Users {
		var acc orm.Account
		status, _ := verify.Status(worldPerspective, serviceID, userMetadata.Id)

		userState := VoiceUserState{
			Timestamp:          ts,
			ServiceID:          serviceID,
			ServiceUserID:      userMetadata.Id,
			ChannelID:          channelID,
			Muted:              userMetadata.Muted,
			Deafened:           userMetadata.Deafened,
			WvWRank:            acc.WvWRank,
			Age:                int(acc.Age),
			VerificationStatus: status.Status.ID(),
		}
		tx.NewInsert().Model(&userState).Exec(ctx)
	}

	// Commit transaction
	err = tx.Commit()
	if err != nil {
		return errors.WithStack(err)
	}
	committed = true
	return nil
}
