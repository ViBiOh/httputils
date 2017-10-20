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
