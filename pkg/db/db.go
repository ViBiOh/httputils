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

type key int

const (
	ctxTxKey key = iota
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
	if len(host) == 0 {
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

// StoreTx stores given transaction in context
func StoreTx(ctx context.Context, tx *sql.Tx) context.Context {
	return context.WithValue(ctx, ctxTxKey, tx)
}

// ReadTx read transaction stored in context
func ReadTx(ctx context.Context) *sql.Tx {
	value := ctx.Value(ctxTxKey)
	if value == nil {
		return nil
	}

	if tx, ok := value.(*sql.Tx); ok {
		return tx
	}

	return nil
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
func GetRow(ctx context.Context, db *sql.DB, query string, args ...interface{}) *sql.Row {
	tx := ReadTx(ctx)

	ctx, cancel := context.WithTimeout(ctx, SQLTimeout)
	defer cancel()

	if tx != nil {
		return tx.QueryRowContext(ctx, query, args...)
	}
	return db.QueryRowContext(ctx, query, args...)
}

// Create execute query with a RETURNING id
func Create(ctx context.Context, db *sql.DB, query string, args ...interface{}) (newID uint64, err error) {
	tx := ReadTx(ctx)
	if tx == nil {
		if tx, err = db.Begin(); err != nil {
			return
		}

		defer func() {
			err = EndTx(tx, err)
		}()
	}

	ctx, cancel := context.WithTimeout(ctx, SQLTimeout)
	defer cancel()

	err = tx.QueryRowContext(ctx, query, args...).Scan(&newID)
	return
}

// Exec execute query with specified timeout, disregarding result
func Exec(ctx context.Context, db *sql.DB, query string, args ...interface{}) (err error) {
	tx := ReadTx(ctx)
	if tx == nil {
		if tx, err = db.Begin(); err != nil {
			return
		}

		defer func() {
			err = EndTx(tx, err)
		}()
	}

	ctx, cancel := context.WithTimeout(ctx, SQLTimeout)
	defer cancel()

	_, err = tx.ExecContext(ctx, query, args...)
	return
}
