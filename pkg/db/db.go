package db

import (
	"database/sql"
	"flag"
	"fmt"
	"strings"

	"github.com/ViBiOh/httputils/v2/pkg/errors"
	"github.com/ViBiOh/httputils/v2/pkg/tools"
	_ "github.com/lib/pq" // Not referenced but needed for database/sql
)

// Config of package
type Config struct {
	host *string
	port *string
	user *string
	pass *string
	name *string
}

// Flags adds flags for configuring package
func Flags(fs *flag.FlagSet, prefix string) Config {
	return Config{
		host: tools.NewFlag(prefix, "database").Name("Host").Default("").Label("Host").ToString(fs),
		port: tools.NewFlag(prefix, "database").Name("Port").Default("5432").Label("Port").ToString(fs),
		user: tools.NewFlag(prefix, "database").Name("User").Default("").Label("User").ToString(fs),
		pass: tools.NewFlag(prefix, "database").Name("Pass").Default("").Label("Pass").ToString(fs),
		name: tools.NewFlag(prefix, "database").Name("Name").Default("").Label("Name").ToString(fs),
	}
}

// New creates new App from Config
func New(config Config) (*sql.DB, error) {
	host := strings.TrimSpace(*config.host)
	if host == "" {
		return nil, nil
	}

	port := strings.TrimSpace(*config.port)
	user := strings.TrimSpace(*config.user)
	pass := strings.TrimSpace(*config.pass)
	name := strings.TrimSpace(*config.name)

	db, err := sql.Open("postgres", fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", host, port, user, pass, name))
	if err != nil {
		return nil, errors.WithStack(err)
	}

	if err = db.Ping(); err != nil {
		return db, errors.WithStack(err)
	}

	return db, nil
}

// Ping indicate if database is ready or not
func Ping(db *sql.DB) bool {
	return db != nil && db.Ping() == nil
}

// GetTx return given transaction if not nil or create a new one
func GetTx(db *sql.DB, tx *sql.Tx) (*sql.Tx, error) {
	if db != nil && tx == nil {
		usedTx, err := db.Begin()

		if err != nil {
			return nil, errors.WithStack(err)
		}
		return usedTx, nil
	}

	return tx, nil
}

// EndTx ends transaction according error without shadowing given error
func EndTx(tx *sql.Tx, err error) error {
	if err != nil {
		if endErr := tx.Rollback(); endErr != nil {
			return errors.New("%#v, and also %#v", err, endErr)
		}
	} else if endErr := tx.Commit(); endErr != nil {
		return errors.WithStack(err)
	}

	return err
}

// RowsClose closes rows without shadowing error
func RowsClose(rows *sql.Rows, err error) error {
	if endErr := rows.Close(); endErr != nil {
		if err == nil {
			return errors.WithStack(endErr)
		}
		return errors.New("%#v, and also %#v", err, endErr)
	}

	return err
}
