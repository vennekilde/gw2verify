package verify

import (
	"time"

	"github.com/vennekilde/gw2apidb/pkg/orm"
	"github.com/vennekilde/gw2verify/internal/config"

	"github.com/vennekilde/gw2apidb/pkg/gw2api"
)

type VerificationStatusExt struct {
	Status    VerificationStatus
	Expires   int
	AccountID string
}

type VerificationStatus int

const (
	ACCESS_DENIED_UNKNOWN                 VerificationStatus = 0
	ACCESS_GRANTED_HOME_WORLD             VerificationStatus = 1
	ACCESS_GRANTED_LINKED_WORLD           VerificationStatus = 2
	ACCESS_GRANTED_HOME_WORLD_TEMPORARY   VerificationStatus = 3
	ACCESS_GRANTED_LINKED_WORLD_TEMPORARY VerificationStatus = 4
	ACCESS_DENIED_ACCOUNT_NOT_LINKED      VerificationStatus = 5
	ACCESS_DENIED_EXPIRED                 VerificationStatus = 6
	ACCESS_DENIED_INVALID_WORLD           VerificationStatus = 7
	ACCESS_DENIED_BANNED                  VerificationStatus = 8
)

const (
	None          string = "None"
	PlayForFree   string = "PlayForFree"
	GuildWars2    string = "GuildWars2"
	HeartOfThorns string = "HeartOfThorns"
	PathOfFire    string = "PathOfFire"
)

func (status VerificationStatus) Name() string {
	switch status {
	case 0:
		return "ACCESS_DENIED_UNKNOWN"
	case 1:
		return "ACCESS_GRANTED_HOME_WORLD"
	case 2:
		return "ACCESS_GRANTED_LINKED_WORLD"
	case 3:
		return "ACCESS_GRANTED_HOME_WORLD_TEMPORARY"
	case 4:
		return "ACCESS_GRANTED_LINKED_WORLD_TEMPORARY"
	case 5:
		return "ACCESS_DENIED_ACCOUNT_NOT_LINKED"
	case 6:
		return "ACCESS_DENIED_EXPIRED"
	case 7:
		return "ACCESS_DENIED_INVALID_WORLD"
	case 8:
		return "ACCESS_DENIED_BANNED"
	default:
		return "UNKNOWN"
	}
}

func (status VerificationStatus) AccessGranted() bool {
	return status >= 1 || status <= 4
}

func (status VerificationStatus) AccessDenied() bool {
	return status.AccessGranted() == false
}

type ServiceLink struct {
	gw2api.Gw2Model
	AccountID              string
	ServiceID              int    `gorm:"primary_key:true"`
	ServiceUserID          string `gorm:"primary_key:true"`
	IsPrimary              bool
	ServiceUserDisplayName string
}

type Configuration struct {
	LinkedWorlds []int
}

var Config = Configuration{
	LinkedWorlds: []int{2013},
}

func Status(serviceID int, serviceUserID string) (status VerificationStatusExt) {
	return StatusWithAccount(serviceID, serviceUserID, nil)
}
func StatusWithAccount(serviceID int, serviceUserID string, accData *gw2api.Account) (status VerificationStatusExt) {
	if serviceUserID == "" {
		status.Status = ACCESS_DENIED_ACCOUNT_NOT_LINKED
		return status
	}

	//Check if user is linked to a gw2 account
	link := ServiceLink{}
	err := orm.DB().First(&link, "service_id = ? AND service_user_id = ?", serviceID, serviceUserID).Error
	if link.AccountID != "" {
		//Check verification status of linked account
		acc := gw2api.Account{}
		if accData == nil {
			err = orm.DB().First(&acc, "id = ?", link.AccountID).Error
		} else {
			acc = *accData
		}
		if err == nil {
			status = AccountStatus(acc)
			//Return status if access has been granted, or if the user is banned
			if status.Status.AccessGranted() || status.Status == ACCESS_DENIED_BANNED {
				return status
			}
		}
	}

	tempAccess := TemporaryAccess{}
	err = orm.DB().First(&tempAccess, "service_id = ? AND service_user_id = ?", serviceID, serviceUserID).Error
	if err == nil && tempAccess.ServiceUserID != "" {
		timeSinceGranted := int(time.Since(tempAccess.DbUpdated).Seconds())
		status.Expires = config.Config().TemporaryAccessExpirationTime - timeSinceGranted
		if timeSinceGranted >= config.Config().TemporaryAccessExpirationTime {
			status.Status = ACCESS_DENIED_EXPIRED
			return status
		}
		if tempAccess.World == config.Config().HomeWorld {
			status.Status = ACCESS_GRANTED_HOME_WORLD_TEMPORARY
			return status
		}
		for _, world := range Config.LinkedWorlds {
			if world == tempAccess.World {
				status.Status = ACCESS_GRANTED_LINKED_WORLD_TEMPORARY
				return status
			}
		}
		status.Status = ACCESS_DENIED_INVALID_WORLD
		return status
	}
	status.Status = ACCESS_DENIED_ACCOUNT_NOT_LINKED
	return status
}

func AccountStatus(acc gw2api.Account) (status VerificationStatusExt) {
	if acc.ID == "" {
		status.Status = ACCESS_DENIED_ACCOUNT_NOT_LINKED
		return status
	}
	status.AccountID = acc.ID
	//Ban logic
	if false {
		status.Status = ACCESS_DENIED_BANNED
		return status
	}

	if int(time.Since(acc.DbUpdated).Seconds()) >= config.Config().ExpirationTime {
		status.Status = ACCESS_DENIED_EXPIRED
		return status
	}

	if acc.World == config.Config().HomeWorld {
		status.Status = ACCESS_GRANTED_HOME_WORLD
		return status
	}
	for _, world := range Config.LinkedWorlds {
		if world == acc.World {
			status.Status = ACCESS_GRANTED_LINKED_WORLD
			return status
		}
	}

	status.Status = ACCESS_DENIED_INVALID_WORLD
	return status
}
