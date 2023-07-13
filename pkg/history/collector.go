package history

import (
	"context"
	"database/sql"
	"strconv"
	"time"

	"github.com/vennekilde/gw2verify/internal/orm"
	"gitlab.com/MrGunflame/gw2api"
	"go.uber.org/zap"
)

type History struct {
	RID       int64 `bun:"r_id,pk,scanonly"`
	Type      HistoryType
	AccountID string
	Timestamp time.Time `bun:",scanonly"`
	Old       sql.NullString
	New       sql.NullString
}

type HistoryType string

const (
	WorldMove  HistoryType = "WorldMove"
	Registered HistoryType = "Registered"
)

func Collect() error {
	ctx := context.Background()

	db := orm.DB()
	tokens := []gw2api.TokenInfo{}
	err := db.NewSelect().Model(&tokens).Where("'progression'=ANY(permissions)").Scan(ctx)
	if err != nil {
		return err
	}

	zap.L().Info("tokens with progression permission", zap.Int("count", len(tokens)))

	return nil
}

func CollectAccount(storedAcc orm.Account, acc gw2api.Account) error {
	ctx := context.Background()
	db := orm.DB()
	if storedAcc.World != acc.World {
		event := &History{
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
		}
		_, err := db.NewInsert().Model(event).Exec(ctx)
		if err != nil {
			zap.L().Warn("unable to store world move event", zap.Any("event", event), zap.Error(err))
		}
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
