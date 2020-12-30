package verify

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/golang/glog"
	"github.com/jinzhu/gorm"

	"github.com/vennekilde/gw2apidb/pkg/gw2api"
	"github.com/vennekilde/gw2apidb/pkg/orm"
	"github.com/vennekilde/gw2verify/internal/api/types"
	"github.com/vennekilde/gw2verify/internal/config"
	"github.com/vennekilde/gw2verify/pkg/utils"
)

// FreeToPlayWvWRankRestriction restricts the minimum required wvw rank before free to play users can verify
var FreeToPlayWvWRankRestriction = 0

type VerifictionStatusListener struct {
	WorldPerspective int
	ServiceID        int
	Listener         chan types.VerificationStatus
}

var ServicePollListeners map[int]VerifictionStatusListener = make(map[int]VerifictionStatusListener)

// SetAPIKeyByUserService sets an apikey from a user of a specific service
func SetAPIKeyByUserService(gw2API *gw2api.GW2Api, worldPerspective int, serviceID int, serviceUserID string, primary bool, apikey string, ignoreRestrictions bool) (err error, userErr error) {
	//Stip spaces
	apikey = utils.StripWhitespace(apikey)

	if err = gw2API.SetAuthenticationWithoutCheck(apikey, []string{"account"}); err != nil {
		return fmt.Errorf("Could not set APIKey '%s' for user %s on service %d. Error: %#v", apikey, serviceUserID, serviceID, err), err
	}

	token, err := gw2API.TokenInfo()
	if err != nil {
		return fmt.Errorf("Could not set APIKey '%s' for user %s on service %d. Error: %#v", apikey, serviceUserID, serviceID, err), err
	}
	acc, err := gw2API.Account()
	if err != nil {
		return fmt.Errorf("Could not set APIKey '%s' for user %s on service %d. Error: %#v", apikey, serviceUserID, serviceID, err), err
	}

	// Ensure no one us already linked with this account
	var storedAcc gw2api.Account
	if err = orm.DB().First(&storedAcc, "id = ?", acc.ID).Error; err != nil && err != gorm.ErrRecordNotFound {
		return err, nil
	}

	//Set additional token permissions
	if err = gw2API.SetAuthenticationWithoutCheck(apikey, token.Permissions); err != nil {
		return fmt.Errorf("Could not set APIKey '%s' for user %s on service %d. Error: %#v", apikey, serviceUserID, serviceID, err), err
	}

	// Check if apikey & account pass restrictions
	if ignoreRestrictions == false {
		// Ensure no one is already linked with this account
		var storedLink ServiceLink
		if err = orm.DB().First(&storedLink, "service_id = ? AND account_id = ? AND is_primary = true", serviceID, acc.ID).Error; err != nil && err != gorm.ErrRecordNotFound {
			return err, nil
		}
		if storedLink.ServiceUserID != "" && storedLink.ServiceUserID != serviceUserID {
			return fmt.Errorf("Could not set APIKey '%s' for user %s on service %d. Already linked with another user with id %s", apikey, serviceUserID, serviceID, storedLink.ServiceUserID), errors.New("Account already linked with another user. Contact an admin if you need to transfer access to your current discord user")
		}

		// Additional restrictions
		err = processRestrictions(gw2API, worldPerspective, acc, token, serviceID, serviceUserID)
		if err != nil {
			return err, err
		}
	}

	// Check if a notification should be published
	CheckForVerificationUpdate(storedAcc, acc)

	err = token.Persist(gw2API.Auth, acc.ID)
	if err != nil {
		return fmt.Errorf("Could not persist tokeninfo information: APIKey '%s' for user %s on service %d. Error: %#v", apikey, serviceUserID, serviceID, err), nil
	}
	glog.Infof("Stored APIKey '%s' with account id: %s", apikey, acc.ID)

	err = acc.Persist()
	if err != nil {
		return fmt.Errorf("Could not persist account information: APIKey '%s' for user %s on service %d. Error: %#v", apikey, serviceUserID, serviceID, err), nil
	}

	userErr = SetOrReplaceServiceLink(serviceID, serviceUserID, primary, acc.ID)
	return nil, userErr
}

func processRestrictions(gw2api *gw2api.GW2Api, worldPerspective int, acc gw2api.Account, token gw2api.TokenInfo, serviceID int, serviceUserID string) (err error) {
	if config.Config().SkipRestrictions {
		return nil
	}
	if err := processAPIKeyRestrictions(worldPerspective, acc, token, serviceID, serviceUserID); err != nil {
		return err
	}
	if err := processAccountRestrictions(worldPerspective, acc); err != nil {
		return err
	}
	if err := processCharacterRestrictions(gw2api, acc); err != nil {
		return err
	}
	return nil
}

func processAPIKeyRestrictions(worldPerspective int, acc gw2api.Account, token gw2api.TokenInfo, serviceID int, serviceUserID string) (err error) {
	//Check if api key is named correctly
	apiKeyCode := GetAPIKeyCode(serviceID, serviceUserID)
	if strings.Contains(strings.ToUpper(token.Name), apiKeyCode) == false {
		return fmt.Errorf("APIKey name incorrect. You need to name your api key \"%s\" instead of \"%s\"", GetAPIKeyName(worldPerspective, serviceID, serviceUserID), token.Name)
	}

	freeToPlay := IsFreeToPlay(acc)

	//FreeToPlay restrictions
	if freeToPlay {
		//Ensure progression permission is present
		hasProgression := Contains(token.Permissions, "progression")

		//Ensure characters permission is present
		hasCharacters := Contains(token.Permissions, "characters")

		if !hasProgression || !hasCharacters {
			return errors.New("Missing apikey permission \"characters\" and or \"progression\".\nYou are trying to verify a FreeToPlay account and is therefore required to have a level 80 character")
		}
	}

	return err
}

func processAccountRestrictions(worldPerspective int, acc gw2api.Account) (err error) {
	freeToPlay := IsFreeToPlay(acc)

	//FreeToPlay restrictions
	if freeToPlay {
		//Check if FreeToPlay player meets WvW rank requirements
		if acc.WvWRank < FreeToPlayWvWRankRestriction {
			return fmt.Errorf("You need to have WvW rank %d to verify yourself. This is required for all FreeToPlay accounts\nCurrent WvW rank: %d", FreeToPlayWvWRankRestriction, acc.WvWRank)
		}
	}
	return err
}

func processCharacterRestrictions(gw2api *gw2api.GW2Api, acc gw2api.Account) (err error) {
	freeToPlay := IsFreeToPlay(acc)

	//FreeToPlay restrictions
	if freeToPlay {
		//Fetch character names
		charNames, err := gw2api.Characters()
		if err != nil {
			return err
		}

		//Fetch characters
		chars, err := gw2api.CharacterIds(charNames...)
		if err != nil {
			return err
		}

		//Calc highest char level
		highestLevel := 0
		for _, char := range chars {
			if char.Level > highestLevel {
				highestLevel = char.Level
			}
		}

		//FreeToPlay level 80 retrictions
		if highestLevel < 80 {
			return fmt.Errorf("You need to be level 80 to verify yourself. This is required for all FreeToPlay accounts\nHighest character level found: %d", highestLevel)
		}

	}
	return err
}

// CheckForVerificationUpdate checks if the verification status for the user has changed since last synchronization
// If it has, it will send out a notification to all registered listeners
func CheckForVerificationUpdate(storedAcc gw2api.Account, acc gw2api.Account) (err error) {
	if storedAcc.World != acc.World || int(time.Since(storedAcc.DbUpdated).Seconds()) >= config.Config().ExpirationTime {
		err = OnVerificationUpdate(acc)
	}
	return err
}

// OnVerificationUpdate sends out a verification notification to all registered listeners
func OnVerificationUpdate(acc gw2api.Account) (err error) {
	links := []ServiceLink{}
	if err = orm.DB().Find(&links, "account_id = ?", acc.ID).Error; err != nil {
		return err
	}

	for _, link := range links {
		serviceListener := ServicePollListeners[link.ServiceID]
		if serviceListener.Listener != nil {
			acc.DbUpdated = time.Now().UTC()
			status, _, err := StatusWithAccount(serviceListener.WorldPerspective, link.ServiceID, link.ServiceUserID, &acc)
			if err != nil {
				glog.Error(err)
				continue
			}
			verificationStatus := types.VerificationStatus{
				Account_id: status.AccountData.ID,
				Expires:    status.Expires,
				Status:     types.EnumVerificationStatusStatus(status.Status.Name()),
				Service_links: []types.ServiceLink{
					{
						Display_name:    link.ServiceUserDisplayName,
						Service_id:      link.ServiceID,
						Service_user_id: link.ServiceUserID,
					},
				},
			}
			serviceListener.Listener <- verificationStatus
		}
	}

	return err
}

// SetOrReplaceServiceLink creates or replaces a service link between a service user and an account
func SetOrReplaceServiceLink(serviceID int, serviceUserID string, primary bool, accountID string) (err error) {
	link := ServiceLink{}
	if primary {
		qResult := orm.DB().Where("service_id = ? AND account_id = ? AND is_primary = true", serviceID, accountID).Delete(&link)
		if qResult.Error != nil {
			return qResult.Error
		}
		if qResult.RowsAffected > 0 {
			glog.Infof("Removed %d rows while replacing service link {serviceID: %d, serviceUserID: %s, accountID: %s, primary: true}", qResult.RowsAffected, serviceID, serviceUserID, accountID)
		}
	}
	link.ServiceUserID = serviceUserID
	link.ServiceID = serviceID
	link.AccountID = accountID
	link.IsPrimary = primary
	//link.ServiceUserDisplayName = ""
	if err := orm.DB().Omit("db_created").Save(&link).Error; err != nil {
		return fmt.Errorf("Could not persist service link: User %s on service %d. Error: %#v", serviceUserID, serviceID, err)
	}
	glog.Infof("Stored service link {ServiceID: %d, ServiceUserID: %s, AccountID: %s}", link.ServiceID, link.ServiceUserID, link.ServiceUserID)
	return err
}

// Contains checks if a slice contains the given item
func Contains(slice []string, item string) bool {
	for _, itemInSlice := range slice {
		if itemInSlice == item {
			return true
		}
	}
	return false
}

// IsFreeToPlay returns true if the account is a free to play account
func IsFreeToPlay(acc gw2api.Account) bool {
	return Contains(acc.Access, PlayForFree) && !Contains(acc.Access, GuildWars2)
}
