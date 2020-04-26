package db

import (
	"context"
	"database/sql"
	"errors"
	"flag"
	"fmt"
	"strings"
	"time"

	"github.com/ViBiOh/httputils/v3/pkg/flags"
	_ "github.com/lib/pq" // Not referenced but needed for database/sql
)

var (
	// ErrNoHost occurs when host is not provided in configuration
	ErrNoHost = errors.New("no host for database connection")

	// SQLTimeout when running queries
	SQLTimeout = time.Second * 5
)

// Config of package
type Config struct {
	host    *string
	port    *uint
	user    *string
	pass    *string
	name    *string
	sslmode *string
}

// Flags adds flags for configuring package
func Flags(fs *flag.FlagSet, prefix string) Config {
	return Config{
		host:    flags.New(prefix, "database").Name("Host").Default("").Label("Host").ToString(fs),
		port:    flags.New(prefix, "database").Name("Port").Default(5432).Label("Port").ToUint(fs),
		user:    flags.New(prefix, "database").Name("User").Default("").Label("User").ToString(fs),
		pass:    flags.New(prefix, "database").Name("Pass").Default("").Label("Pass").ToString(fs),
		name:    flags.New(prefix, "database").Name("Name").Default("").Label("Name").ToString(fs),
		sslmode: flags.New(prefix, "database").Name("Sslmode").Default("disable").Label("SSL Mode").ToString(fs),
	}
}

// New creates new App from Config
func New(config Config) (*sql.DB, error) {
	host := strings.TrimSpace(*config.host)
	if host == "" {
		return nil, ErrNoHost
	}

	user := strings.TrimSpace(*config.user)
	pass := strings.TrimSpace(*config.pass)
	name := strings.TrimSpace(*config.name)
	sslmode := strings.TrimSpace(*config.sslmode)

	db, err := sql.Open("postgres", fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s", host, *config.port, user, pass, name, sslmode))
	if err != nil {
		return nil, err
	}

	if err = db.Ping(); err != nil {
		return db, err
	}

	return db, nil
}

// Ping indicate if database is ready or not
func Ping(db *sql.DB) bool {
	ctx, cancel := context.WithTimeout(context.Background(), SQLTimeout)
	defer cancel()

	return db != nil && db.PingContext(ctx) == nil
}

// GetTx return given transaction if not nil or create a new one
func GetTx(db *sql.DB, tx *sql.Tx) (*sql.Tx, error) {
	if db == nil || tx != nil {
		return tx, nil
	}

	return db.Begin()
}

// EndTx ends transaction according error without shadowing given error
func EndTx(tx *sql.Tx, err error) error {
	if err == nil {
		return tx.Commit()
	}

	if rollbackErr := tx.Rollback(); rollbackErr != nil {
		return fmt.Errorf("%s: %w", err.Error(), rollbackErr)
	}

	return err
}

// RowsClose closes rows without shadowing error
func RowsClose(rows *sql.Rows, err error) error {
	if closeErr := rows.Close(); closeErr != nil {
		if err == nil {
			return closeErr
		}

		return fmt.Errorf("%s: %w", err.Error(), closeErr)
	}

	return err
}

// GetRow execute single row query
func GetRow(db *sql.DB, query string, args ...interface{}) *sql.Row {
	ctx, cancel := context.WithTimeout(context.Background(), SQLTimeout)
	defer cancel()

	return db.QueryRowContext(ctx, query, args...)
}

// Create execute query with a RETURNING id
func Create(db *sql.DB, tx *sql.Tx, query string, args ...interface{}) (newID uint64, err error) {
	var usedTx *sql.Tx
	if usedTx, err = GetTx(db, tx); err != nil {
		return
	}

	if usedTx != tx {
		defer func() {
			err = EndTx(usedTx, err)
		}()
	}

	ctx, cancel := context.WithTimeout(context.Background(), SQLTimeout)
	defer cancel()

	err = usedTx.QueryRowContext(ctx, query, args...).Scan(&newID)
	return
}

// Exec execute query with specified timeout, disregarding result
func Exec(db *sql.DB, tx *sql.Tx, query string, args ...interface{}) (err error) {
	var usedTx *sql.Tx
	if usedTx, err = GetTx(db, tx); err != nil {
		return
	}

	if usedTx != tx {
		defer func() {
			err = EndTx(usedTx, err)
		}()
	}

	ctx, cancel := context.WithTimeout(context.Background(), SQLTimeout)
	defer cancel()

	_, err = usedTx.ExecContext(ctx, query, args...)
	return
}
