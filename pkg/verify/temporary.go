package verify

import (
	"time"

	"github.com/vennekilde/gw2apidb/pkg/gw2api"
	"github.com/vennekilde/gw2apidb/pkg/orm"
)

type TemporaryAccess struct {
	gw2api.Gw2Model
	ServiceID     int    `gorm:"unique_index:idx_ta_service_id_service_user_id"`
	ServiceUserID string `gorm:"unique_index:idx_ta_service_id_service_user_id"`
	World         int
}

func GrantTemporaryWorldAssignment(serviceID int, serviceUserID string, world int) (err error, userErr error) {
	temporaryAccess := TemporaryAccess{
		ServiceUserID: serviceUserID,
		ServiceID:     serviceID,
		World:         world,
	}
	temporaryAccess.DbUpdated = time.Now().UTC()
	return orm.DB().Assign(temporaryAccess).FirstOrCreate(&temporaryAccess).Error, nil
}
