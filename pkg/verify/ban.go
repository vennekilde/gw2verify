package verify

import (
	"context"
	"time"

	"github.com/vennekilde/gw2verify/internal/api"
	"github.com/vennekilde/gw2verify/internal/orm"
	"go.uber.org/zap"
)

// GetBan returns the longest active ban on an account, if they have any
func GetBan(acc *orm.Account) *orm.Ban {
	ctx := context.Background()
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
func BanServiceUser(expiration time.Time, reason string, serviceID int, serviceUserID string) error {
	ctx := context.Background()

	// Extract user id from service user information
	link, err := orm.GetServiceLink(serviceID, serviceUserID)
	if err != nil {
		return err
	}
	// Format ban
	ban := orm.Ban{
		UserID: link.UserID,
		BanData: api.BanData{
			Until:  expiration,
			Reason: reason,
		},
	}
	// Persist ban
	_, err = orm.DB().NewInsert().
		Model(&ban).
		Exec(ctx)

	if err != nil {
		return err
	}

	// Gather all accounts associated with the now banned user
	accounts, err := orm.GetUserAccounts(link.UserID)
	if err != nil {
		return err
	}

	//  Notify all listeners of ban
	for _, account := range accounts {
		err := OnVerificationUpdate(account)
		if err != nil {
			zap.L().Error("unable to perform verification update after ban",
				zap.Any("account", account),
				zap.Error(err),
			)
		}
	}

	return nil
}
