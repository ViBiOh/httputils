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
	"github.com/lib/pq"
	_ "github.com/lib/pq" // Not referenced but needed for database/sql
)

type key int

const (
	ctxTxKey key = iota
)

var (
	// ErrNoHost occurs when host is not provided in configuration
	ErrNoHost = errors.New("no host for database connection")

	// ErrNoTransaction occurs when no transaction is provided but needed
	ErrNoTransaction = errors.New("no transaction in context, please wrap with DoAtomic()")

	// ErrBulkEnded occurs when bulk creation is over
	ErrBulkEnded = errors.New("no more data to copy")

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
func Flags(fs *flag.FlagSet, prefix string, overrides ...flags.Override) Config {
	return Config{
		host:    flags.New(prefix, "database").Name("Host").Default(flags.Default("Host", "", overrides)).Label("Host").ToString(fs),
		port:    flags.New(prefix, "database").Name("Port").Default(flags.Default("Port", 5432, overrides)).Label("Port").ToUint(fs),
		user:    flags.New(prefix, "database").Name("User").Default(flags.Default("User", "", overrides)).Label("User").ToString(fs),
		pass:    flags.New(prefix, "database").Name("Pass").Default(flags.Default("Pass", "", overrides)).Label("Pass").ToString(fs),
		name:    flags.New(prefix, "database").Name("Name").Default(flags.Default("Name", "", overrides)).Label("Name").ToString(fs),
		sslmode: flags.New(prefix, "database").Name("Sslmode").Default(flags.Default("Sslmode", "disable", overrides)).Label("SSL Mode").ToString(fs),
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

func readTx(ctx context.Context) *sql.Tx {
	value := ctx.Value(ctxTxKey)
	if value == nil {
		return nil
	}

	if tx, ok := value.(*sql.Tx); ok {
		return tx
	}

	return nil
}

// DoAtomic execute given action in a transactionnal context
func DoAtomic(ctx context.Context, db *sql.DB, action func(context.Context) error) error {
	if action == nil {
		return errors.New("no action provided")
	}

	if readTx(ctx) != nil {
		return action(ctx)
	}

	tx, err := db.Begin()
	if err != nil {
		return err
	}

	err = action(StoreTx(ctx, tx))
	if err == nil {
		return tx.Commit()
	}

	if rollbackErr := tx.Rollback(); rollbackErr != nil {
		return fmt.Errorf("%s: %w", err.Error(), rollbackErr)
	}

	return err
}

// List execute multiple rows query
func List(ctx context.Context, db *sql.DB, scanner func(*sql.Rows) error, query string, args ...interface{}) error {
	ctx, cancel := context.WithTimeout(ctx, SQLTimeout)
	defer cancel()

	var rows *sql.Rows
	var err error

	if tx := readTx(ctx); tx != nil {
		rows, err = tx.QueryContext(ctx, query, args...)
	} else if db != nil {
		rows, err = db.QueryContext(ctx, query, args...)
	} else {
		return errors.New("no transaction or database provided")
	}

	if err != nil {
		return err
	}

	for rows.Next() && err == nil {
		err = scanner(rows)
	}

	closeErr := rows.Close()
	if err == nil {
		return closeErr
	}

	if closeErr == nil {
		return err
	}

	return fmt.Errorf("%s: %w", err.Error(), closeErr)
}

// Get execute single row query
func Get(ctx context.Context, db *sql.DB, scanner func(*sql.Row) error, query string, args ...interface{}) error {
	ctx, cancel := context.WithTimeout(ctx, SQLTimeout)
	defer cancel()

	if tx := readTx(ctx); tx != nil {
		return scanner(tx.QueryRowContext(ctx, query, args...))
	} else if db != nil {
		return scanner(db.QueryRowContext(ctx, query, args...))
	}

	return errors.New("no transaction or database provided")
}

// Create execute query with a RETURNING id
func Create(ctx context.Context, query string, args ...interface{}) (uint64, error) {
	tx := readTx(ctx)
	if tx == nil {
		return 0, ErrNoTransaction
	}

	ctx, cancel := context.WithTimeout(ctx, SQLTimeout)
	defer cancel()

	var newID uint64
	return newID, tx.QueryRowContext(ctx, query, args...).Scan(&newID)
}

// Exec execute query with specified timeout, disregarding result
func Exec(ctx context.Context, query string, args ...interface{}) error {
	tx := readTx(ctx)
	if tx == nil {
		return ErrNoTransaction
	}

	ctx, cancel := context.WithTimeout(ctx, SQLTimeout)
	defer cancel()

	_, err := tx.ExecContext(ctx, query, args...)
	return err
}

// Bulk load data into schema and table by batch
func Bulk(ctx context.Context, feeder func(*sql.Stmt) error, schema, table string, columns ...string) error {
	tx := readTx(ctx)
	if tx == nil {
		return ErrNoTransaction
	}

	stmt, err := tx.Prepare(pq.CopyInSchema(schema, table, columns...))
	if err != nil {
		return fmt.Errorf("unable to prepare context: %s", err)
	}

	for err == nil {
		err = feeder(stmt)
	}

	if err != ErrBulkEnded {
		return fmt.Errorf("unable to feed bulk creation: %s", err)
	}

	ctx, cancel := context.WithTimeout(ctx, SQLTimeout)
	defer cancel()

	if _, err := stmt.ExecContext(ctx); err != nil {
		return fmt.Errorf("unable to exec bulk creation: %s", err)
	}

	if err := stmt.Close(); err != nil {
		return fmt.Errorf("unable to close bulk creation: %s", err)
	}

	return nil
}
