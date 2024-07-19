package verify

import (
	"context"
	"time"

	"github.com/vennekilde/gw2verify/v2/internal/api"
	"github.com/vennekilde/gw2verify/v2/internal/orm"
)

type BanService struct {
	em *EventEmitter
}

func NewBanService(em *EventEmitter) *BanService {
	return &BanService{
		em: em,
	}
}

// GetActiveBan returns the longest active ban on an account, if they have any
func GetActiveBan(bans []api.Ban) *api.Ban {
	for i := range bans {
		ban := &bans[i]
		if time.Now().Before(ban.Until) {
			return ban
		}
	}
	return nil
}

// GetBan returns the longest active ban on an account, if they have any
func GetBan(acc *orm.Account) *orm.Ban {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	//Find Longest active ban
	ban := orm.Ban{}
	result := orm.DB().NewSelect().
		Model(&ban).
		Where("user_id = ? AND until > NOW()", acc.UserID).
		// Limit to longest lasting ban
		Order("until desc").
		Limit(1).
		Scan(ctx)
	if result != nil {
		return nil
	}
	return &ban
}

// BanServiceUser bans a user's gw2 account for the given duration
func (bs *BanService) BanServiceUser(expiration time.Time, reason string, platformID int, platformUserId string) error {
	// Extract user id from service user information
	link, err := orm.GetPlatformLink(platformID, platformUserId)
	if err != nil {
		return err
	}

	return bs.BanUser(expiration, reason, link.UserID)
}

// BanServiceUser bans a user's gw2 account for the given duration
func (bs *BanService) BanUser(expiration time.Time, reason string, userID int64) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Format ban
	ban := orm.Ban{
		Ban: api.Ban{
			UserID: userID,
			Until:  expiration,
			Reason: reason,
		},
	}
	// Persist ban
	_, err := orm.DB().NewInsert().
		Model(&ban).
		Exec(ctx)

	if err != nil {
		return err
	}

	// Gather all accounts associated with the now banned user
	var user api.User
	err = orm.QueryGetUser(orm.DB(), &user, userID).
		Model(&user).
		Scan(ctx)
	if err != nil {
		return err
	}

	bs.em.Process(nil, &user)
	return nil
}
