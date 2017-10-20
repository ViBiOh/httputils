package db

import (
	"database/sql"
	"fmt"

	// Not referenced but needed for database/sql
	_ "github.com/lib/pq"
)

// GetDB start DB connection
func GetDB(dbHost, dbPort, dbUser, dbPass, dbName string) (*sql.DB, error) {
	if dbHost == `` {
		return nil, nil
	}

	db, err := sql.Open(`postgres`, fmt.Sprintf(`host=%s port=%s user=%s password=%s dbname=%s sslmode=disable`, dbHost, dbPort, dbUser, dbPass, dbName))
	if err != nil {
		return nil, fmt.Errorf(`Error while opening database connection: %v`, err)
	}

	if err = db.Ping(); err != nil {
		return nil, fmt.Errorf(`Error while pinging database: %v`, err)
	}

	return db, nil
}

// Ping indicate if database is ready or not
func Ping(db *sql.DB) bool {
	return db != nil && db.Ping() == nil
}

// GetTx return given transaction if not nil or create a new one
func GetTx(db *sql.DB, label string, tx *sql.Tx) (*sql.Tx, error) {
	if tx == nil {
		usedTx, err := db.Begin()

		if err != nil {
			return nil, fmt.Errorf(`Error while getting transaction for %s: %v`, label, err)
		}
		return usedTx, nil
	}

	return tx, nil
}

// EndTx ends transaction according error without shadowing given error
func EndTx(label string, tx *sql.Tx, err error) error {
	if err != nil {
		if endErr := tx.Rollback(); endErr != nil {
			return fmt.Errorf(`%v, and also error while rolling back transaction for %s: %v`, err, label, endErr)
		}
	} else if endErr := tx.Commit(); endErr != nil {
		return fmt.Errorf(`Error while committing transaction for %s: %v`, label, endErr)
	}

	return nil
}

// RowsClose closes rows without shadowing error
func RowsClose(label string, rows *sql.Rows, err error) error {
	if endErr := rows.Close(); endErr != nil {
		endErr = fmt.Errorf(`Error while closing rows for %s: %v`, label, endErr)

		if err == nil {
			return endErr
		}
		return fmt.Errorf(`%v, and also %v`, err, endErr)
	}

	return err
}
