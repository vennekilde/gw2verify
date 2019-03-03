package verify

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/golang/glog"

	"github.com/vennekilde/gw2apidb/pkg/gw2api"
	"github.com/vennekilde/gw2apidb/pkg/orm"
	"github.com/vennekilde/gw2verify/internal/api/handlers/updates"
	"github.com/vennekilde/gw2verify/internal/api/types"
	"github.com/vennekilde/gw2verify/internal/config"
)

func SetAPIKeyByUserService(gw2api *gw2api.GW2Api, serviceID int, serviceUserID string, primary bool, apikey string, ignoreRestrictions bool) (err error, userErr error) {
	if err = gw2api.SetAuthenticationWithoutCheck(apikey, []string{"account"}); err != nil {
		return fmt.Errorf("Could not set APIKey '%s' for user %s on service %d. Error: %#v", apikey, serviceUserID, serviceID, err), err
	}

	token, err := gw2api.TokenInfo()
	if err != nil {
		return fmt.Errorf("Could not set APIKey '%s' for user %s on service %d. Error: %#v", apikey, serviceUserID, serviceID, err), err
	}
	acc, err := gw2api.Account()
	if err != nil {
		return fmt.Errorf("Could not set APIKey '%s' for user %s on service %d. Error: %#v", apikey, serviceUserID, serviceID, err), err
	}

	if ignoreRestrictions == false {
		err = processRestrictions(gw2api, acc, token, serviceID, serviceUserID)
		if err != nil {
			return err, err
		}
	}

	CheckForVerificationUpdate(acc)

	err = token.Persist(gw2api.Auth, acc.ID)
	if err != nil {
		return fmt.Errorf("Could not persist tokeninfo information: APIKey '%s' for user %s on service %d. Error: %#v", apikey, serviceUserID, serviceID, err), nil
	}
	glog.Infof("Stored APIKey '%s' with account id: %s", apikey, acc.ID)

	err = acc.Persist()
	if err != nil {
		return fmt.Errorf("Could not persist account information: APIKey '%s' for user %s on service %d. Error: %#v", apikey, serviceUserID, serviceID, err), nil
	}

	return nil, SetOrReplaceServiceLink(serviceID, serviceUserID, primary, acc.ID)
}

func processRestrictions(gw2api *gw2api.GW2Api, acc gw2api.Account, token gw2api.TokenInfo, serviceID int, serviceUserID string) (err error) {

	if config.Config().SkipRestrictions {
		return nil
	}

	//Check if api key is named correctly
	apiKeyCode := GetAPIKeyCode(serviceID, serviceUserID)
	if strings.Contains(strings.ToUpper(token.Name), apiKeyCode) == false {
		return fmt.Errorf("APIKey name incorrect. You need to name your api key \"%s\" instead of \"%s\"", GetAPIKeyName(serviceID, serviceUserID), token.Name)
	}

	freeToPlay := Contains(acc.Access, PlayForFree)

	//FreeToPlay restrictions
	if freeToPlay {
		//Ensure characters permission is present
		hasCharacters := Contains(token.Permissions, "characters")
		if !hasCharacters {
			return errors.New("Missing apikey permission \"characters\".\nYou are trying to verify a FreeToPlay account and is therefore required to have a level 80 character")
		}

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
			highestLevel = char.Level
		}

		//FreeToPlay level 80 retrictions
		if highestLevel < 80 {
			return fmt.Errorf("You need to be level 80 to verify yourself. This is required for all FreeToPlay accounts\nHighest character level found: %d", highestLevel)
		}

	}
	return err
}

func CheckForVerificationUpdate(acc gw2api.Account) (err error) {
	storedAcc := gw2api.Account{}
	if err = orm.DB().First(&storedAcc, "id = ?", acc.ID).Error; err != nil && err.Error() != "record not found" {
		return err
	}
	if storedAcc.World != acc.World {
		err = OnVerificationUpdate(acc)
	}
	return err
}

func OnVerificationUpdate(acc gw2api.Account) (err error) {
	links := []ServiceLink{}
	if err = orm.DB().Find(&links, "account_id = ?", acc.ID).Error; err != nil {
		return err
	}

	for _, link := range links {
		channel := updates.ServicePollListeners[link.ServiceID]
		if channel != nil {
			acc.DbUpdated = time.Now().UTC()
			status := StatusWithAccount(link.ServiceID, link.ServiceUserID, &acc)
			verificationStatus := types.VerificationStatus{
				Account_id: status.AccountID,
				Expires:    status.Expires,
				Status:     types.EnumVerificationStatusStatus(status.Status.Name()),
				Service_links: []types.ServiceLink{
					types.ServiceLink{
						Display_name:    link.ServiceUserDisplayName,
						Service_id:      link.ServiceID,
						Service_user_id: link.ServiceUserID,
					},
				},
			}
			channel <- verificationStatus
		}
	}

	return err
}

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
	glog.Infof("Stored service link %#v", link)
	return err
}

func Contains(slice []string, item string) bool {
	for _, itemInSlice := range slice {
		if itemInSlice == item {
			return true
		}
	}
	return false
}
