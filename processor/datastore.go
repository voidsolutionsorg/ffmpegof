package processor

import (
	"embed"
	"fmt"
	"time"

	"database/sql"
	_ "github.com/lib/pq"
	_ "modernc.org/sqlite"

	"github.com/aleksasiriski/rffmpeg-go/migrate"
)

type datastore struct {
	*sql.DB
	dbType string
}

var now = time.Now

var (
	//go:embed migrations/sqlite
	migrationsSqlite embed.FS
)

var (
	//go:embed migrations/postgres
	migrationsPostgres embed.FS
)

func newDatastore(db *sql.DB, dbType string, mg *migrate.Migrator) (*datastore, error) {
	switch dbType {
	case "sqlite":
		{
			// migrations/sqlite
			if err := mg.Migrate(&migrationsSqlite, "processor"); err != nil {
				return nil, fmt.Errorf("migrate: %w", err)
			}
		}
	case "postgres":
		{
			// migrations/postgres
			if err := mg.Migrate(&migrationsPostgres, "processor"); err != nil {
				return nil, fmt.Errorf("migrate: %w", err)
			}
		}
	default:
		panic(fmt.Errorf("Incorrect database type!"))
	}
	return &datastore{db, dbType}, nil
}
