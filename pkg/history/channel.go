package history

import (
	"time"

	"github.com/vennekilde/gw2apidb/pkg/gw2api"
	"github.com/vennekilde/gw2apidb/pkg/orm"
	"github.com/vennekilde/gw2verify/internal/api/types"
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
	WvWRank            int `json:"wvw_rank"`
	Age                int
	VerificationStatus int
}

func CollectChannelStatistics(serviceID int, channelID string, worldPerspective int, data types.ChannelMetadata) error {
	db := orm.DB()
	tx := db.Begin()
	ts := time.Now()
	for _, userMetadata := range data.Users {
		var acc gw2api.Account
		status, _, _ := verify.StatusWithAccount(worldPerspective, serviceID, userMetadata.Id, &acc)

		userState := VoiceUserState{
			Timestamp:          ts,
			ServiceID:          serviceID,
			ServiceUserID:      userMetadata.Id,
			ChannelID:          channelID,
			Muted:              userMetadata.Muted,
			Deafened:           userMetadata.Deafened,
			WvWRank:            acc.WvWRank,
			Age:                acc.Age,
			VerificationStatus: int(status.Status),
		}
		tx.Save(&userState)
	}
	err := tx.Commit().Error
	return err
}
