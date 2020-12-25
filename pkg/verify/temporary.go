package verify

import (
	"time"

	"github.com/vennekilde/gw2apidb/pkg/gw2api"
	"github.com/vennekilde/gw2apidb/pkg/orm"
)

// TemporaryAccess represents a user that has been granted temporary access from a given world
type TemporaryAccess struct {
	gw2api.Gw2Model
	ServiceID     int    `gorm:"primary_key:true"`
	ServiceUserID string `gorm:"type:varchar(64);primary_key:true"`
	World         int    `gorm:"not null"`
}

// GrantTemporaryWorldAssignment temporarily mark the account as being from the given world
// This will grant the user temporary access, if they are set to be a from a world that would normally grant access
func GrantTemporaryWorldAssignment(serviceID int, serviceUserID string, world int) (err error, userErr error) {
	temporaryAccess := TemporaryAccess{
		ServiceUserID: serviceUserID,
		ServiceID:     serviceID,
		World:         world,
	}
	temporaryAccess.DbUpdated = time.Now().UTC()
	return orm.DB().Assign(temporaryAccess).FirstOrCreate(&temporaryAccess).Error, nil
}
