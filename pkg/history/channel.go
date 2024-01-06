package history

import (
	"context"
	"time"

	"github.com/pkg/errors"
	"github.com/vennekilde/gw2verify/v2/internal/api"
	"github.com/vennekilde/gw2verify/v2/internal/orm"
	"github.com/vennekilde/gw2verify/v2/pkg/verify"
	"go.uber.org/zap"
)

type VoiceUserState struct {
	Timestamp          time.Time
	PlatformID         int
	PlatformUserID     string
	ChannelID          string
	Muted              bool
	Deafened           bool
	WvWRank            int `json:"wvw_rank" bun:"wvw_rank"`
	Age                int
	VerificationStatus int
}

type Statistics struct {
	verificationModule *verify.Verification
}

func NewStatistics(verification *verify.Verification) *Statistics {
	return &Statistics{
		verificationModule: verification,
	}
}

func (s *Statistics) WorldStatistics(platformID int, channelID string, worldPerspective int, data api.ChannelMetadata) error {
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
		var user api.User
		err = orm.QueryGetPlatformUser(tx, &user, platformID, userMetadata.Id).
			Model(&user).
			Scan(ctx)
		if err != nil {
			zap.L().Error("error while fetching user data", zap.Error(err))
			continue
		}
		status := s.verificationModule.Status(worldPerspective, &user)

		userState := VoiceUserState{
			Timestamp:          ts,
			PlatformID:         platformID,
			PlatformUserID:     userMetadata.Id,
			ChannelID:          channelID,
			Muted:              userMetadata.Muted,
			Deafened:           userMetadata.Deafened,
			VerificationStatus: status.ID(),
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
