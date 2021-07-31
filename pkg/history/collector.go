package history

import (
	"database/sql"
	"strconv"
	"time"

	"github.com/vennekilde/gw2apidb/pkg/gw2api"
	"github.com/vennekilde/gw2apidb/pkg/orm"
	"go.uber.org/zap"
)

type History struct {
	RID       uint32         `gorm:"auto_increment;not null;primary_key"`
	Type      HistoryType    `gorm:"size:16;index;not null"`
	AccountID string         `gorm:"type varchar(64);not null;index"`
	Timestamp time.Time      `gorm:"not null;default:CURRENT_TIMESTAMP"`
	Old       sql.NullString `gorm:"size:64"`
	New       sql.NullString `gorm:"size:64"`
}

type HistoryType string

const (
	WorldMove  HistoryType = "WorldMove"
	Registered HistoryType = "Registered"
)

func Collect() error {

	db := orm.DB()
	tokens := []gw2api.TokenInfo{}
	if err := db.Table(gw2api.TableNameTokenInfo).Where("'progression'=ANY(permissions)").Find(&tokens).Error; err != nil {
		return err
	}

	zap.L().Info(len(tokens))

	return nil
}

func CollectAccount(storedAcc gw2api.Account, acc gw2api.Account) error {
	db := orm.DB()
	if storedAcc.World != acc.World {
		db.Save(&History{
			AccountID: acc.ID,
			Type:      WorldMove,
			Old: sql.NullString{
				String: strconv.Itoa(storedAcc.World),
				Valid:  storedAcc.World != 0,
			},
			New: sql.NullString{
				String: strconv.Itoa(acc.World),
				Valid:  true,
			},
		})
	}
	if storedAcc.ID != "" {
		if storedAcc.WvWRank > acc.WvWRank {
			MarkPlaying(acc)
		} else {
			MarkNotPlaying(acc)
		}
	}
	return nil
}

func MarkPlaying(acc gw2api.Account) {
	//Not implemented
}

func MarkNotPlaying(acc gw2api.Account) {
	//Not implemented
}
