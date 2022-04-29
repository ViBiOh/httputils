package db

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"strings"
	"time"

	"github.com/ViBiOh/flags"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"go.opentelemetry.io/otel/trace"
)

type key struct{}

var ctxTxKey key

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
	Query(context.Context, string, ...any) (pgx.Rows, error)
	QueryRow(context.Context, string, ...any) pgx.Row
}

// App of package
type App struct {
	tracer trace.Tracer
	db     Database
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
		host:    flags.String(fs, prefix, "database", "Host", "Host", "", overrides),
		port:    flags.Uint(fs, prefix, "database", "Port", "Port", 5432, overrides),
		user:    flags.String(fs, prefix, "database", "User", "User", "", overrides),
		pass:    flags.String(fs, prefix, "database", "Pass", "Pass", "", overrides),
		name:    flags.String(fs, prefix, "database", "Name", "Name", "", overrides),
		maxConn: flags.Uint(fs, prefix, "database", "MaxConn", "Max Open Connections", 5, overrides),
		sslmode: flags.String(fs, prefix, "database", "Sslmode", "SSL Mode", "disable", overrides),
		timeout: flags.Uint(fs, prefix, "database", "Timeout", "Connect timeout", 10, overrides),
	}
}

// New creates new App from Config
func New(config Config, tracer trace.Tracer) (App, error) {
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
		db:     db,
		tracer: tracer,
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
func (a App) DoAtomic(ctx context.Context, action func(context.Context) error) error {
	if action == nil {
		return errors.New("no action provided")
	}

	if a.tracer != nil {
		var span trace.Span
		ctx, span = a.tracer.Start(ctx, "transaction")
		defer span.End()
	}

	if readTx(ctx) != nil {
		return action(ctx)
	}

	tx, err := a.db.Begin(ctx)
	if err != nil {
		return err
	}

	err = action(StoreTx(ctx, tx))

	if err == nil {
		return tx.Commit(ctx)
	}

	if rollbackErr := tx.Rollback(ctx); rollbackErr != nil {
		return fmt.Errorf("%s: %w", err, rollbackErr)
	}

	return err
}

// List execute multiple rows query
func (a App) List(ctx context.Context, scanner func(pgx.Rows) error, query string, args ...any) (err error) {
	if a.tracer != nil {
		var span trace.Span
		ctx, span = a.tracer.Start(ctx, "list")
		defer span.End()
	}

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
func (a App) Get(ctx context.Context, scanner func(pgx.Row) error, query string, args ...any) error {
	if a.tracer != nil {
		var span trace.Span
		ctx, span = a.tracer.Start(ctx, "get")
		defer span.End()
	}

	ctx, cancel := context.WithTimeout(ctx, SQLTimeout)
	defer cancel()

	if tx := readTx(ctx); tx != nil {
		return scanner(tx.QueryRow(ctx, query, args...))
	}

	return scanner(a.db.QueryRow(ctx, query, args...))
}

// Create execute query with a RETURNING id
func (a App) Create(ctx context.Context, query string, args ...any) (uint64, error) {
	if a.tracer != nil {
		var span trace.Span
		ctx, span = a.tracer.Start(ctx, "create")
		defer span.End()
	}

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
func (a App) Exec(ctx context.Context, query string, args ...any) error {
	_, err := a.exec(ctx, query, args...)
	return err
}

// One execute query with specified timeout, for exactly one row
func (a App) One(ctx context.Context, query string, args ...any) error {
	output, err := a.exec(ctx, query, args...)
	if err != nil {
		return err
	}

	if count := output.RowsAffected(); count != 1 {
		return fmt.Errorf("%d rows affected, wanted 1", count)
	}

	return nil
}

func (a App) exec(ctx context.Context, query string, args ...any) (pgconn.CommandTag, error) {
	if a.tracer != nil {
		var span trace.Span
		ctx, span = a.tracer.Start(ctx, "exec")
		defer span.End()
	}

	tx := readTx(ctx)
	if tx == nil {
		return nil, ErrNoTransaction
	}

	ctx, cancel := context.WithTimeout(ctx, SQLTimeout)
	defer cancel()

	return tx.Exec(ctx, query, args...)
}

type feeder struct {
	err     error
	fetcher func() ([]any, error)
	values  []any
}

func (bc *feeder) Next() bool {
	bc.values, bc.err = bc.fetcher()
	return bc.err == nil && len(bc.values) != 0
}

func (bc *feeder) Values() ([]any, error) {
	return bc.values, bc.err
}

func (bc *feeder) Err() error {
	return bc.err
}

// Bulk load data into schema and table by batch
func (a App) Bulk(ctx context.Context, fetcher func() ([]any, error), schema, table string, columns ...string) error {
	if a.tracer != nil {
		var span trace.Span
		ctx, span = a.tracer.Start(ctx, "bulk")
		defer span.End()
	}

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
