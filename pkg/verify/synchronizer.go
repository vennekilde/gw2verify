package verify

import (
	"errors"
	"time"

	"github.com/vennekilde/gw2apidb/pkg/orm"
	"github.com/vennekilde/gw2verify/internal/config"

	"github.com/golang/glog"
	"github.com/vennekilde/gw2apidb/pkg/gw2api"
	"github.com/vennekilde/gw2apidb/pkg/gw2apidb"
)

// StartAPISynchronizer starts a synchronization loop that will continuesly fetch the oldest updated API key
// and synchronize it with the gw2 api
func StartAPISynchronizer(gw2API *gw2api.GW2Api) {
	var attemptsSinceLastSuccess int
	var acc gw2api.Account
	for {
		if attemptsSinceLastSuccess >= 10 {
			time.Sleep(10 * time.Second)
		}
		tokeninfo, err := gw2apidb.FindLastUpdatedAPIKey(config.Config().ExpirationTime)
		if err != nil {
			glog.Errorf("Could not retrieve APIKey from storage. Error: %#v", err)
			attemptsSinceLastSuccess++
			continue
		}
		if tokeninfo.APIKey == "" {
			glog.Errorf("Retrieved APIKey from storage is empty. Data: %#v", tokeninfo)
			attemptsSinceLastSuccess++
			continue
		}

		tokeninfo.APIKey = SpaceStringsBuilder(tokeninfo.APIKey)

		acc, err = SynchronizeAPIKey(gw2API, tokeninfo.APIKey, tokeninfo.Permissions)
		if err != nil {
			goto SyncError
		} else {
			if config.Config().Debug {
				glog.Infof("Updated account: %s", acc.Name)
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
			glog.Errorf("Could not synchronize apikey '%s' for account '%s'. Error: %s", tokeninfo.APIKey, acc.Name, err)
		} else {
			glog.Errorf("Could not synchronize apikey '%s'. Error: %s", tokeninfo.APIKey, err)
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

	CheckForVerificationUpdate(acc)

	acc.Persist()

	return acc, err
}

func SynchronizeLinkedUser(gw2apiclient *gw2api.GW2Api, serviceID int, serviceUserID string) (err error, userErr error) {
	link := ServiceLink{}
	err = orm.DB().First(&link, "service_id = ? AND service_user_id = ?", serviceID, serviceUserID).Error
	if err != nil {
		return err, nil
	}
	if link.AccountID == "" {
		err = errors.New("No service link found with that user id and service id")
		return err, err
	}

	tokeninfo := gw2api.TokenInfo{}
	err = orm.DB().First(&tokeninfo, "account_id = ?", link.AccountID).Error
	if err != nil {
		return err, nil
	}
	if tokeninfo.AccountID != link.AccountID {
		err = errors.New("No apikey associated with found service link")
		return err, err
	}

	_, err = SynchronizeAPIKey(gw2apiclient, tokeninfo.APIKey, tokeninfo.Permissions)
	return nil, err
}
