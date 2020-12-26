package verify

import (
	"strconv"
	"time"

	"github.com/golang/glog"
	"github.com/jinzhu/gorm"
	"github.com/vennekilde/gw2apidb/pkg/orm"
	"github.com/vennekilde/gw2verify/internal/config"

	"github.com/vennekilde/gw2apidb/pkg/gw2api"
)

type VerificationStatusExt struct {
	Status      VerificationStatus
	Expires     int64
	AccountData gw2api.Account
	ServiceLink ServiceLink
	Description string
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
	ACCESS_DENIED_REQUIREMENT_NOT_MET     VerificationStatus = 9
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
	case 9:
		return "ACCESS_DENIED_REQUIREMENT_NOT_MET"
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
	AccountID              string `gorm:"type varchar(64);not null"`
	ServiceID              int    `gorm:"primary_key:true"`
	ServiceUserID          string `gorm:"primary_key:true"`
	IsPrimary              bool   `gorm:"not null"`
	ServiceUserDisplayName string `gorm:"varchar(128);not null"`
}

type Configuration struct {
	LinkedWorlds map[int][]int
}

func Status(worldPerspective int, serviceID int, serviceUserID string) (status VerificationStatusExt, link ServiceLink) {
	return StatusWithAccount(worldPerspective, serviceID, serviceUserID, nil)
}
func StatusWithAccount(worldPerspective int, serviceID int, serviceUserID string, accData *gw2api.Account) (status VerificationStatusExt, link ServiceLink) {
	var err error
	if serviceUserID == "" {
		status.Status = ACCESS_DENIED_ACCOUNT_NOT_LINKED
		return status, link
	}

	//Check if user is linked to a gw2 account
	if err = orm.DB().First(&link, "service_id = ? AND service_user_id = ?", serviceID, serviceUserID).Error; err != nil {
		if err != gorm.ErrRecordNotFound {
			glog.Error(err)
		}
		status.Status = ACCESS_DENIED_UNKNOWN
		return status, link
	}
	if link.AccountID != "" {
		//Check verification status of linked account
		acc := gw2api.Account{}
		if accData == nil {
			err = orm.DB().First(&acc, "id = ?", link.AccountID).Error
		} else {
			acc = *accData
		}
		if err == nil {
			status = AccountStatus(acc, worldPerspective)
			//Return status if access has been granted, or if the user is banned
			if status.Status.AccessGranted() || status.Status == ACCESS_DENIED_BANNED {
				return status, link
			}
		}
	}

	tempAccess := TemporaryAccess{}
	if err = orm.DB().First(&tempAccess, "service_id = ? AND service_user_id = ?", serviceID, serviceUserID).Error; err != nil {
		if err != gorm.ErrRecordNotFound {
			glog.Error(err)
		}
		status.Status = ACCESS_DENIED_UNKNOWN
		return status, link
	}
	if tempAccess.ServiceUserID != "" {
		timeSinceGranted := int(time.Since(tempAccess.DbUpdated).Seconds())
		status.Expires = int64(config.Config().TemporaryAccessExpirationTime - timeSinceGranted)
		if timeSinceGranted >= config.Config().TemporaryAccessExpirationTime {
			status.Status = ACCESS_DENIED_EXPIRED
			return status, link
		}
		if tempAccess.World == worldPerspective {
			status.Status = ACCESS_GRANTED_HOME_WORLD_TEMPORARY
			return status, link
		}
		//Get cached world links
		worldLinks, err := GetWorldLinks(worldPerspective)
		if err != nil {
			if err != gorm.ErrRecordNotFound {
				glog.Error(err)
			}
			status.Status = ACCESS_DENIED_UNKNOWN
			return status, link
		}
		for _, world := range worldLinks {
			if world == tempAccess.World {
				status.Status = ACCESS_GRANTED_LINKED_WORLD_TEMPORARY
				return status, link
			}
		}
		status.Status = ACCESS_DENIED_INVALID_WORLD
		return status, link
	}
	status.Status = ACCESS_DENIED_ACCOUNT_NOT_LINKED
	return status, link
}

// AccountStatus checks the verification access status for an account given a world perspective
func AccountStatus(acc gw2api.Account, worldPerspective int) (status VerificationStatusExt) {
	if acc.ID == "" {
		status.Status = ACCESS_DENIED_ACCOUNT_NOT_LINKED
		return status
	}
	status.AccountData = acc
	//Ban logic
	banStatus := GetBan(acc)
	if banStatus != nil {
		status.Status = ACCESS_DENIED_BANNED
		status.Expires = banStatus.Expires.Sub(time.Now()).Milliseconds()
		status.Description = "Banned until " + banStatus.Expires.String() + " \nReason: " + banStatus.Reason
		return status
	}

	//Check if access is expired
	if int(time.Since(acc.DbUpdated).Seconds()) >= config.Config().ExpirationTime {
		//Check if last time token is expired
		tokens := []gw2api.TokenInfo{}
		//For some reason, the Go pg library complains with "pq: got 2 parameters but the statement requires 1"
		//when the param is inside the ' ' closure, so we insert it as part of the string instead
		//It's an int anyway and we trust the source, so it should be fine
		orm.DB().Find(&tokens, "account_id = ? AND last_success >= db_updated - interval '"+strconv.Itoa(config.Config().ExpirationTime)+" seconds'", acc.ID)
		if len(tokens) <= 0 {
			//No valid api keys found, therefore must be expired
			status.Status = ACCESS_DENIED_EXPIRED
			return status
		}
	}

	if err := processAccountRestrictions(worldPerspective, acc); err != nil {
		status.Status = ACCESS_DENIED_REQUIREMENT_NOT_MET
		status.Description = err.Error()
		return status
	}

	if acc.World == worldPerspective {
		status.Status = ACCESS_GRANTED_HOME_WORLD
		return status
	}

	//Get cached world links
	worldLinks, err := GetWorldLinks(worldPerspective)
	if err != nil {
		if err != gorm.ErrRecordNotFound {
			glog.Error(err)
		}
		status.Status = ACCESS_DENIED_UNKNOWN
		return status
	}
	for _, world := range worldLinks {
		if world == acc.World {
			status.Status = ACCESS_GRANTED_LINKED_WORLD
			return status
		}
	}

	status.Status = ACCESS_DENIED_INVALID_WORLD
	return status
}
