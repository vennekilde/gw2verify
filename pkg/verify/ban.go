package verify

import (
	"time"

	"github.com/vennekilde/gw2apidb/pkg/gw2api"
	"github.com/vennekilde/gw2apidb/pkg/orm"
)

// Ban contains information on length and why an account was banned
type Ban struct {
	gw2api.Gw2Model
	AccountID string    `json:"account_id" gorm:"type:varchar(64)"`
	Expires   time.Time `gorm:"default:3000-01-01 00:00:00.000000+00"`
	Reason    string    `json:"reason" gorm:"type:text"`
}

// GetBan returns the longest active ban on an accountd, if they have any
func GetBan(acc gw2api.Account) *Ban {
	//Find Longest active ban
	ban := Ban{}
	result := orm.DB().Order("expires desc").First(&ban, "account_id = ? AND expires > NOW()", acc.ID)
	if result.Error != nil {
		return nil
	}
	return &ban
}
