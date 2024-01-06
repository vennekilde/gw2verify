package orm

import (
	"fmt"

	"github.com/uptrace/bun"
	"github.com/vennekilde/gw2verify/v2/internal/config"
)

func QueryGetUser(idb bun.IDB, model any, userID int64) *bun.SelectQuery {
	query := QueryGetUsers(idb, model).
		Where("\"user\".id = ?", userID)
	return query
}

func QueryGetUsers(idb bun.IDB, model any) *bun.SelectQuery {
	query := idb.NewSelect().
		Model(model).
		Relation("Bans").
		Relation("Accounts", func(sq *bun.SelectQuery) *bun.SelectQuery {
			return sq.Where(fmt.Sprintf("NOW() < db_updated + interval '%d seconds'", config.Config().ExpirationTime))
		}).
		Relation("PlatformLinks").
		Relation("EphemeralAssociations", func(sq *bun.SelectQuery) *bun.SelectQuery {
			return sq.Where("NOW() < until")
		})
	return query
}

func QueryGetPlatformUser(idb bun.IDB, model any, platformID int, platformUserID string) *bun.SelectQuery {
	query := QueryGetPlatformUsers(idb, model).
		Where("platform_link.platform_id = ? AND platform_link.platform_user_id = ?", platformID, platformUserID)

	return query
}

func QueryGetPlatformUsers(idb bun.IDB, model any) *bun.SelectQuery {
	query := QueryGetUsers(idb, model).
		Join("INNER JOIN platform_links as platform_link ON \"user\".id = platform_link.user_id")

	return query
}
