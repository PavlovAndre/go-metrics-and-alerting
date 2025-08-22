package migrations

import (
	"database/sql"
	"github.com/PavlovAndre/go-metrics-and-alerting.git/internal/logger"
	"github.com/lopezator/migrator"
)

func Migrations() (*migrator.Migrator, error) {
	// Configure migrations

	m, err := migrator.New(
		migrator.Migrations(&migrator.Migration{
			Name: "Create metric tables",
			Func: func(tx *sql.Tx) error {
				logger.Log.Infow("database migrations started")
				if _, err := tx.Exec("CREATE TABLE metrics (name VARCHAR PRIMARY KEY, type TEXT NOT NULL, value double precision, delta bigint);"); err != nil {
					return err
				}
				logger.Log.Infow("database migrations finished")
				/*if _, err := tx.Exec("CREATE TABLE t_gauge (name VARCHAR PRIMARY KEY, value double precision);"); err != nil {
					return err
				}
				if _, err := tx.Exec("CREATE TABLE t_counter (name VARCHAR PRIMARY KEY, value bigint);"); err != nil {
					return err
				}*/
				return nil
			},
		}))
	return m, err
}
