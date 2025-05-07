package fix

import (
	"database/sql"
	gerrors "github.com/go-errors/errors"
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/constancy/migration"
	"github.com/uptrace/bun"
)

func Migration24(db *bun.DB) error {
	//check migration in table "bun_migrations"
	migrationExist, err := checkMigration24Exist(db)
	if err != nil {
		return gerrors.Wrap(err, 0)
	}
	if migrationExist {
		//check column "tls" in table "clusters"
		tlsExist, err := checkTls(db)
		if err != nil {
			return gerrors.Wrap(err, 0)
		}
		if !tlsExist {
			logger.Warnf("Detected absence of clusters.tls of migration # 24. Applying...")
			_, err := db.Exec(`ALTER TABLE clusters ADD COLUMN IF NOT EXISTS tls jsonb;`)
			if err != nil {
				return gerrors.Wrap(err, 0)
			}
			logger.Warnf("Migration #24 has been applied")
		}
	}
	return nil
}

func checkMigration24Exist(db *bun.DB) (bool, error) {
	var name string
	err := db.QueryRow(`
		SELECT name FROM ? WHERE name = '00000000000024' LIMIT 1
	`, bun.Safe(migrationTableName)).Scan(&name)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, err
	}
	return name != "", nil
}

func checkTls(db *bun.DB) (bool, error) {
	result, err := db.Query(`SELECT column_name 
		FROM information_schema.columns 
		WHERE table_name='clusters' and column_name='tls'`,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, err
	}
	rowsCount := migration.CountRows(result)
	if rowsCount > 1 {
		panic("Returned more than 1 rows")
	} else {
		return true, nil
	}
}
