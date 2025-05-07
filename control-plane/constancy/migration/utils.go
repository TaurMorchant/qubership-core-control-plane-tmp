package migration

import (
	"database/sql"
	"embed"
	"github.com/netcracker/qubership-core-lib-go/v3/logging"
	"github.com/uptrace/bun/migrate"
)

var log = logging.GetLogger("pg-migration")

var migrations = migrate.NewMigrations()

//go:embed *.sql
var sqlMigrations embed.FS

func GetMigrations() (*migrate.Migrations, error) {
	if err := migrations.Discover(sqlMigrations); err != nil {
		log.Errorf("Can't find sql migrations %+v", err.Error())
		return nil, err
	}
	return migrations, nil
}

func contains(arr []v7Route, route v7Route) bool {
	for _, elem := range arr {
		if elem == route {
			return true
		}
	}
	return false
}

func CountRows(rows *sql.Rows) int {
	count := 0
	for rows.Next() {
		var name string
		err := rows.Scan(&name)
		if err != nil {
			break
		}
		count++
	}
	return count
}
