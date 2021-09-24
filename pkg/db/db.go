package db

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"strings"
	"time"

	"github.com/ViBiOh/httputils/v4/pkg/flags"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
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

// Database interface needed for working
//go:generate mockgen -destination ../mocks/database.go -mock_names Database=Database -package mocks github.com/ViBiOh/httputils/v4/pkg/db Database
type Database interface {
	Ping(context.Context) error
	Close()
	Begin(context.Context) (pgx.Tx, error)
	Query(context.Context, string, ...interface{}) (pgx.Rows, error)
	QueryRow(context.Context, string, ...interface{}) pgx.Row
}

// App of package
type App struct {
	db Database
}

// Config of package
type Config struct {
	host    *string
	port    *uint
	user    *string
	pass    *string
	name    *string
	sslmode *string
	maxConn *uint
	timeout *uint
}

// Flags adds flags for configuring package
func Flags(fs *flag.FlagSet, prefix string, overrides ...flags.Override) Config {
	return Config{
		host:    flags.New(prefix, "database", "Host").Default("", overrides).Label("Host").ToString(fs),
		port:    flags.New(prefix, "database", "Port").Default(5432, overrides).Label("Port").ToUint(fs),
		user:    flags.New(prefix, "database", "User").Default("", overrides).Label("User").ToString(fs),
		pass:    flags.New(prefix, "database", "Pass").Default("", overrides).Label("Pass").ToString(fs),
		name:    flags.New(prefix, "database", "Name").Default("", overrides).Label("Name").ToString(fs),
		maxConn: flags.New(prefix, "database", "MaxConn").Default(5, overrides).Label("Max Open Connections").ToUint(fs),
		sslmode: flags.New(prefix, "database", "Sslmode").Default("disable", overrides).Label("SSL Mode").ToString(fs),
		timeout: flags.New(prefix, "database", "Timeout").Default(10, overrides).Label("Connect timeout").ToUint(fs),
	}
}

// New creates new App from Config
func New(config Config) (App, error) {
	host := strings.TrimSpace(*config.host)
	if len(host) == 0 {
		return App{}, ErrNoHost
	}

	user := strings.TrimSpace(*config.user)
	pass := *config.pass
	name := strings.TrimSpace(*config.name)
	sslmode := *config.sslmode

	db, err := pgxpool.Connect(context.Background(), fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s connect_timeout=%d pool_max_conns=%d", host, *config.port, user, pass, name, sslmode, *config.timeout, *config.maxConn))
	if err != nil {
		return App{}, fmt.Errorf("unable to connect to postgres: %s", err)
	}

	instance := App{
		db: db,
	}

	return instance, instance.Ping()
}

// Enabled check if sql.DB is provided
func (a App) Enabled() bool {
	return a.db != nil
}

// Ping indicate if database is ready or not
func (a App) Ping() error {
	ctx, cancel := context.WithTimeout(context.Background(), SQLTimeout)
	defer cancel()

	return a.db.Ping(ctx)
}

// Close the database connection
func (a App) Close() {
	a.db.Close()
}

// StoreTx stores given transaction in context
func StoreTx(ctx context.Context, tx pgx.Tx) context.Context {
	return context.WithValue(ctx, ctxTxKey, tx)
}

func readTx(ctx context.Context) pgx.Tx {
	value := ctx.Value(ctxTxKey)
	if value == nil {
		return nil
	}

	if tx, ok := value.(pgx.Tx); ok {
		return tx
	}

	return nil
}

// DoAtomic execute given action in a transactionnal context
func (a App) DoAtomic(ctx context.Context, action func(context.Context) error) (err error) {
	if action == nil {
		return errors.New("no action provided")
	}

	if readTx(ctx) != nil {
		return action(ctx)
	}

	var tx pgx.Tx

	tx, err = a.db.Begin(ctx)
	if err != nil {
		return
	}

	defer func() {
		if err == nil {
			err = tx.Commit(ctx)
		} else if rollbackErr := tx.Rollback(ctx); rollbackErr != nil {
			err = fmt.Errorf("%s: %w", err.Error(), rollbackErr)
		}
	}()

	err = action(StoreTx(ctx, tx))
	return
}

// List execute multiple rows query
func (a App) List(ctx context.Context, scanner func(pgx.Rows) error, query string, args ...interface{}) (err error) {
	ctx, cancel := context.WithTimeout(ctx, SQLTimeout)
	defer cancel()

	var rows pgx.Rows

	if tx := readTx(ctx); tx != nil {
		rows, err = tx.Query(ctx, query, args...)
	} else {
		rows, err = a.db.Query(ctx, query, args...)
	}

	if err != nil {
		return err
	}

	for rows.Next() && err == nil {
		err = scanner(rows)
	}

	rows.Close()

	return
}

// Get execute single row query
func (a App) Get(ctx context.Context, scanner func(pgx.Row) error, query string, args ...interface{}) error {
	ctx, cancel := context.WithTimeout(ctx, SQLTimeout)
	defer cancel()

	if tx := readTx(ctx); tx != nil {
		return scanner(tx.QueryRow(ctx, query, args...))
	}

	return scanner(a.db.QueryRow(ctx, query, args...))
}

// Create execute query with a RETURNING id
func (a App) Create(ctx context.Context, query string, args ...interface{}) (uint64, error) {
	tx := readTx(ctx)
	if tx == nil {
		return 0, ErrNoTransaction
	}

	ctx, cancel := context.WithTimeout(ctx, SQLTimeout)
	defer cancel()

	var newID uint64
	return newID, tx.QueryRow(ctx, query, args...).Scan(&newID)
}

// Exec execute query with specified timeout, disregarding result
func (a App) Exec(ctx context.Context, query string, args ...interface{}) error {
	tx := readTx(ctx)
	if tx == nil {
		return ErrNoTransaction
	}

	ctx, cancel := context.WithTimeout(ctx, SQLTimeout)
	defer cancel()

	_, err := tx.Exec(ctx, query, args...)
	return err
}

type feeder struct {
	err     error
	fetcher func() ([]interface{}, error)
	values  []interface{}
}

func (bc *feeder) Next() bool {
	bc.values, bc.err = bc.fetcher()
	return bc.err == nil && len(bc.values) != 0
}

func (bc *feeder) Values() ([]interface{}, error) {
	return bc.values, bc.err
}

func (bc *feeder) Err() error {
	return bc.err
}

// Bulk load data into schema and table by batch
func (a App) Bulk(ctx context.Context, fetcher func() ([]interface{}, error), schema, table string, columns ...string) error {
	tx := readTx(ctx)
	if tx == nil {
		return ErrNoTransaction
	}

	ctx, cancel := context.WithTimeout(ctx, SQLTimeout)
	defer cancel()

	if _, err := tx.CopyFrom(ctx, pgx.Identifier{schema, table}, columns, &feeder{fetcher: fetcher}); err != nil {
		return fmt.Errorf("unable to copy from: %s", err)
	}

	return nil
}
