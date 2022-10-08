package db

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"strings"
	"time"

	"github.com/ViBiOh/flags"
	"github.com/ViBiOh/httputils/v4/pkg/tracer"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"go.opentelemetry.io/otel/trace"
)

//go:generate mockgen -source db.go -destination ../mocks/db.go  -package mocks -mock_names Database=Database

type key struct{}

var ctxTxKey key

var (
	ErrNoHost = errors.New("no host for database connection")

	ErrNoTransaction = errors.New("no transaction in context, please wrap with DoAtomic()")

	SQLTimeout = time.Second * 5
)

type Database interface {
	Ping(context.Context) error
	Close()
	Begin(context.Context) (pgx.Tx, error)
	Query(context.Context, string, ...any) (pgx.Rows, error)
	QueryRow(context.Context, string, ...any) pgx.Row
}

type App struct {
	tracer trace.Tracer
	db     Database
}

type Config struct {
	host    *string
	port    *uint
	user    *string
	pass    *string
	name    *string
	sslmode *string
	maxConn *uint
}

func Flags(fs *flag.FlagSet, prefix string, overrides ...flags.Override) Config {
	return Config{
		host:    flags.String(fs, prefix, "database", "Host", "Host", "", overrides),
		port:    flags.Uint(fs, prefix, "database", "Port", "Port", 5432, overrides),
		user:    flags.String(fs, prefix, "database", "User", "User", "", overrides),
		pass:    flags.String(fs, prefix, "database", "Pass", "Pass", "", overrides),
		name:    flags.String(fs, prefix, "database", "Name", "Name", "", overrides),
		maxConn: flags.Uint(fs, prefix, "database", "MaxConn", "Max Open Connections", 5, overrides),
		sslmode: flags.String(fs, prefix, "database", "Sslmode", "SSL Mode", "disable", overrides),
	}
}

func New(config Config, tracer trace.Tracer) (App, error) {
	host := strings.TrimSpace(*config.host)
	if len(host) == 0 {
		return App{}, ErrNoHost
	}

	user := strings.TrimSpace(*config.user)
	pass := *config.pass
	name := strings.TrimSpace(*config.name)
	sslmode := *config.sslmode

	db, err := pgxpool.Connect(context.Background(), fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s pool_max_conns=%d", host, *config.port, user, pass, name, sslmode, *config.maxConn))
	if err != nil {
		return App{}, fmt.Errorf("connect to postgres: %w", err)
	}

	instance := App{
		db:     db,
		tracer: tracer,
	}

	return instance, instance.Ping()
}

func (a App) Enabled() bool {
	return a.db != nil
}

func (a App) Ping() error {
	if !a.Enabled() {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), SQLTimeout)
	defer cancel()

	return a.db.Ping(ctx)
}

func (a App) Close() {
	a.db.Close()
}

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

func (a App) DoAtomic(ctx context.Context, action func(context.Context) error) error {
	if action == nil {
		return errors.New("no action provided")
	}

	ctx, end := tracer.StartSpan(ctx, a.tracer, "transaction", trace.WithSpanKind(trace.SpanKindClient))
	defer end()

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

func (a App) List(ctx context.Context, scanner func(pgx.Rows) error, query string, args ...any) (err error) {
	ctx, end := tracer.StartSpan(ctx, a.tracer, "list", trace.WithSpanKind(trace.SpanKindClient))
	defer end()

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

func (a App) Get(ctx context.Context, scanner func(pgx.Row) error, query string, args ...any) error {
	ctx, end := tracer.StartSpan(ctx, a.tracer, "get", trace.WithSpanKind(trace.SpanKindClient))
	defer end()

	ctx, cancel := context.WithTimeout(ctx, SQLTimeout)
	defer cancel()

	if tx := readTx(ctx); tx != nil {
		return scanner(tx.QueryRow(ctx, query, args...))
	}

	return scanner(a.db.QueryRow(ctx, query, args...))
}

func (a App) Create(ctx context.Context, query string, args ...any) (uint64, error) {
	ctx, end := tracer.StartSpan(ctx, a.tracer, "create", trace.WithSpanKind(trace.SpanKindClient))
	defer end()

	tx := readTx(ctx)
	if tx == nil {
		return 0, ErrNoTransaction
	}

	ctx, cancel := context.WithTimeout(ctx, SQLTimeout)
	defer cancel()

	var newID uint64

	return newID, tx.QueryRow(ctx, query, args...).Scan(&newID)
}

func (a App) Exec(ctx context.Context, query string, args ...any) error {
	_, err := a.exec(ctx, query, args...)

	return err
}

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
	ctx, end := tracer.StartSpan(ctx, a.tracer, "exec", trace.WithSpanKind(trace.SpanKindClient))
	defer end()

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

func (a App) Bulk(ctx context.Context, fetcher func() ([]any, error), schema, table string, columns ...string) error {
	ctx, end := tracer.StartSpan(ctx, a.tracer, "bulk", trace.WithSpanKind(trace.SpanKindClient))
	defer end()

	tx := readTx(ctx)
	if tx == nil {
		return ErrNoTransaction
	}

	ctx, cancel := context.WithTimeout(ctx, SQLTimeout)
	defer cancel()

	if _, err := tx.CopyFrom(ctx, pgx.Identifier{schema, table}, columns, &feeder{fetcher: fetcher}); err != nil {
		return fmt.Errorf("copy from: %w", err)
	}

	return nil
}
