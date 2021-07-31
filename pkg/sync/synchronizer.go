package sync

import (
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/vennekilde/gw2apidb/pkg/orm"
	"github.com/vennekilde/gw2verify/internal/config"
	"github.com/vennekilde/gw2verify/pkg/history"
	"github.com/vennekilde/gw2verify/pkg/utils"
	"github.com/vennekilde/gw2verify/pkg/verify"
	"go.uber.org/zap"

	"github.com/vennekilde/gw2apidb/pkg/gw2api"
	"github.com/vennekilde/gw2apidb/pkg/gw2apidb"
)

// StartAPISynchronizer starts a synchronization loop that will continuesly fetch the oldest updated API key
// and synchronize it with the gw2 api
func StartAPISynchronizer(gw2API *gw2api.GW2Api) {
	var attemptsSinceLastSuccess int
	var acc gw2api.Account
	var successCount = 0
	var successTimestamp = time.Now()
	for {
		if attemptsSinceLastSuccess >= 10 {
			zap.L().Warn("10 consecutive failures, sleeping for 10 seconds")
			time.Sleep(10 * time.Second)
		}
		tokeninfo, err := gw2apidb.FindLastUpdatedAPIKey(config.Config().ExpirationTime)
		if err != nil {
			zap.L().Error("could not retrieve APIKey from storage", zap.Error(err))
			attemptsSinceLastSuccess++
			continue
		}
		if tokeninfo.APIKey == "" {
			zap.L().Error("retrieved APIKey from storage is empty", zap.Any("tokeninfo", tokeninfo))
			attemptsSinceLastSuccess++
			continue
		}

		tokeninfo.APIKey = utils.StripWhitespace(tokeninfo.APIKey)

		acc, err = SynchronizeAPIKey(gw2API, tokeninfo.APIKey, tokeninfo.Permissions)
		if err != nil {
			goto SyncError
		} else {
			if config.Config().Debug {
				zap.L().Info("updated account", zap.string("account name", acc.Name))
			}
			successCount++
			if time.Since(successTimestamp).Minutes() >= 10 {
				zap.L().Info("successful refreshes last 10 minutes", zap.Int("count", successCount))
				successTimestamp = time.Now()
				successCount = 0
			}
			tokeninfo.UpdateLastSuccessfulUpdate()
			attemptsSinceLastSuccess = 0
		}

		//Check if token metadata is missing
		if tokeninfo.AccountID == "" || len(tokeninfo.Permissions) <= 0 {
			//Retrieve tokeninfo from api and persist it
			nTokenInfo, err := gw2API.TokenInfo()
			if err != nil {
				goto SyncError
			}

			if err = nTokenInfo.Persist(tokeninfo.APIKey, acc.ID); err != nil {
				goto SyncError
			}
		}

		//Skip error handling
		continue

	SyncError:
		if acc.Name != "" {
			zap.L().Error("could not synchronize apikey",
				zap.String("apikey", tokeninfo.APIKey),
				zap.String("account name", acc.Name),
				zap.Error(err))
		} else {
			// Show error if in debug mode, or if error is not just an error, stating it is an invalid key
			showErr := !strings.Contains(err.Error(), "invalid key") || config.Config().Debug
			if showErr {
				zap.L().Error("could not synchronize apikey",
					zap.String("apikey", tokeninfo.APIKey),
					zap.Error(err))
			}
		}
		tokeninfo.UpdateLastAttemptedUpdate()
		attemptsSinceLastSuccess++
		continue
	}
}

func SynchronizeAPIKey(gw2API *gw2api.GW2Api, apikey string, permissions []string) (acc gw2api.Account, err error) {
	err = gw2API.SetAuthenticationWithoutCheck(apikey, permissions)
	if err != nil {
		return acc, err
	}
	acc, err = gw2API.Account()
	if err != nil {
		return acc, err
	}

	storedAcc := gw2api.Account{}
	if err = orm.DB().First(&storedAcc, "id = ?", acc.ID).Error; err != nil && err != gorm.ErrRecordNotFound {
		return acc, err
	}

	history.CollectAccount(storedAcc, acc)
	verify.CheckForVerificationUpdate(storedAcc, acc)

	acc.Persist()

	return acc, err
}

func SynchronizeLinkedUser(gw2apiclient *gw2api.GW2Api, serviceID int, serviceUserID string) (err error, userErr error) {
	link := verify.ServiceLink{}
	err = orm.DB().First(&link, "service_id = ? AND service_user_id = ?", serviceID, serviceUserID).Error
	if err != nil {
		return err, nil
	}
	if link.AccountID == "" {
		err = errors.New("no service link found with that user id and service id")
		return err, err
	}

	tokeninfo := gw2api.TokenInfo{}
	err = orm.DB().First(&tokeninfo, "account_id = ?", link.AccountID).Error
	if err != nil {
		return err, nil
	}
	if tokeninfo.AccountID != link.AccountID {
		err = errors.New("no apikey associated with found service link")
		return err, err
	}

	_, err = SynchronizeAPIKey(gw2apiclient, tokeninfo.APIKey, tokeninfo.Permissions)
	return nil, err
}
