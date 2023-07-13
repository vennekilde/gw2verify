package orm

import (
	"database/sql"
	"fmt"
	"sync"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"

	"github.com/vennekilde/gw2verify/internal/config"
)

var db *bun.DB
var m sync.Mutex

func DB() *bun.DB {
	m.Lock()
	defer m.Unlock()
	if db == nil {
		// Build postgres conn string
		dsn := fmt.Sprintf(
			"postgres://%s:%s@%s:%d/%s?sslmode=disable",
			config.Config().PostgresUser,
			config.Config().PostgresPassword,
			config.Config().PostgresHost,
			config.Config().PostgresPort,
			config.Config().PostgresDatabase)

		sqldb := sql.OpenDB(pgdriver.NewConnector(pgdriver.WithDSN(dsn)))

		// Wrap in bun
		bunDB := bun.NewDB(sqldb, pgdialect.New())

		// save singleton
		db = bunDB
	}
	return db
}
