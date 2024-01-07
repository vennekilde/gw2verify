package migrations

import (
	"database/sql"
	"embed"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"go.uber.org/zap"
)

// MigrateDB will perform migration on the configured data using the
// configured postgres server connection from config.GetPostgresDBConfig()
func MigrateDB(db *sql.DB, fs embed.FS) (err error) {
	zap.L().Info("checking migration status")

	// Create schema if not exists
	/*_, err = db.Query("CREATE SCHEMA IF NOT EXISTS " + conf.PostgresSQLSchema)
	if err != nil {
		zap.L().Panic("unable to create missing schema", zap.String("schema", conf.PostgresSQLSchema), zap.Error(err))
	}*/

	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		zap.L().Panic("could not connect to database for migration", zap.Error(err))
	}

	// Load migration files
	d, err := iofs.New(fs, "migrations")
	if err != nil {
		zap.L().Panic("unable to load database migration sql files", zap.Error(err))
	}

	// Prepare migration driver
	m, err := migrate.NewWithInstance(
		"iofs", d,
		"postgres", driver)
	if err != nil {
		zap.L().Panic("could not connect to database for migration", zap.Error(err))
	}

	// Migrate
	err = m.Up()
	if err != nil {
		if err == migrate.ErrNoChange {
			zap.L().Info("no migration performed, already migrated to latest version")
		} else {
			zap.L().Panic("unable to perform database migration", zap.Error(err))
		}
	}

	return nil
}
