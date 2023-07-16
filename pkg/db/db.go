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
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

//go:generate mockgen -source db.go -destination ../mocks/db.go  -package mocks -mock_names Database=Database

type key struct{}

var ctxTxKey key

var (
	ErrNoHost        = errors.New("no host for database connection")
	ErrNoTransaction = errors.New("no transaction in context, please wrap with DoAtomic()")
	SQLTimeout       = time.Second * 5
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
	minConn *uint
	maxConn *uint
}

func Flags(fs *flag.FlagSet, prefix string, overrides ...flags.Override) Config {
	return Config{
		host:    flags.New("Host", "Host").Prefix(prefix).DocPrefix("database").String(fs, "", overrides),
		port:    flags.New("Port", "Port").Prefix(prefix).DocPrefix("database").Uint(fs, 5432, overrides),
		user:    flags.New("User", "User").Prefix(prefix).DocPrefix("database").String(fs, "", overrides),
		pass:    flags.New("Pass", "Pass").Prefix(prefix).DocPrefix("database").String(fs, "", overrides),
		name:    flags.New("Name", "Name").Prefix(prefix).DocPrefix("database").String(fs, "", overrides),
		minConn: flags.New("MinConn", "Min Open Connections").Prefix(prefix).DocPrefix("database").Uint(fs, 2, overrides),
		maxConn: flags.New("MaxConn", "Max Open Connections").Prefix(prefix).DocPrefix("database").Uint(fs, 5, overrides),
		sslmode: flags.New("Sslmode", "SSL Mode").Prefix(prefix).DocPrefix("database").String(fs, "disable", overrides),
	}
}

func New(ctx context.Context, config Config, tracer trace.Tracer) (App, error) {
	host := strings.TrimSpace(*config.host)
	if len(host) == 0 {
		return App{}, ErrNoHost
	}

	user := strings.TrimSpace(*config.user)
	pass := *config.pass
	name := strings.TrimSpace(*config.name)
	sslmode := *config.sslmode

	ctx, cancel := context.WithTimeout(ctx, SQLTimeout)
	defer cancel()

	db, err := pgxpool.New(ctx, fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s pool_min_conns=%d pool_max_conns=%d", host, *config.port, user, pass, name, sslmode, *config.minConn, *config.maxConn))
	if err != nil {
		return App{}, fmt.Errorf("connect to postgres: %w", err)
	}

	instance := App{
		db:     db,
		tracer: tracer,
	}

	return instance, instance.Ping(ctx)
}

func (a App) Enabled() bool {
	return a.db != nil
}

func (a App) Ping(ctx context.Context) error {
	if !a.Enabled() {
		return nil
	}

	ctx, cancel := context.WithTimeout(ctx, SQLTimeout)
	defer cancel()

	return a.db.Ping(ctx)
}

func (a App) Close() {
	if !a.Enabled() {
		return
	}

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

func (a App) DoAtomic(ctx context.Context, action func(context.Context) error) (err error) {
	if action == nil {
		return errors.New("no action provided")
	}

	ctx, end := tracer.StartSpan(ctx, a.tracer, "transaction", trace.WithSpanKind(trace.SpanKindClient))
	defer end(&err)

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

func (a App) Query(ctx context.Context, query string, args ...any) (rows pgx.Rows, err error) {
	ctx, end := tracer.StartSpan(ctx, a.tracer, "query", trace.WithSpanKind(trace.SpanKindClient), trace.WithAttributes(attribute.String("query", query)))
	defer end(&err)

	if tx := readTx(ctx); tx != nil {
		return tx.Query(ctx, query, args...)
	}

	return a.db.Query(ctx, query, args...)
}

func (a App) List(ctx context.Context, scanner func(pgx.Rows) error, query string, args ...any) error {
	ctx, cancel := context.WithTimeout(ctx, SQLTimeout)
	defer cancel()

	rows, err := a.Query(ctx, query, args...)
	if err != nil {
		return err
	}

	for rows.Next() && err == nil {
		err = scanner(rows)
	}

	if readErr := rows.Err(); readErr != nil {
		err = errors.Join(err, readErr)

		rows.Close()
	}

	return err
}

func (a App) QueryRow(ctx context.Context, query string, args ...any) pgx.Row {
	ctx, end := tracer.StartSpan(ctx, a.tracer, "query_row", trace.WithSpanKind(trace.SpanKindClient), trace.WithAttributes(attribute.String("query", query)))
	defer end(nil)

	if tx := readTx(ctx); tx != nil {
		return tx.QueryRow(ctx, query, args...)
	}

	return a.db.QueryRow(ctx, query, args...)
}

func (a App) Get(ctx context.Context, scanner func(pgx.Row) error, query string, args ...any) error {
	ctx, cancel := context.WithTimeout(ctx, SQLTimeout)
	defer cancel()

	return scanner(a.QueryRow(ctx, query, args...))
}

func (a App) Create(ctx context.Context, query string, args ...any) (id uint64, err error) {
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

func (a App) exec(ctx context.Context, query string, args ...any) (command pgconn.CommandTag, err error) {
	ctx, end := tracer.StartSpan(ctx, a.tracer, "exec", trace.WithSpanKind(trace.SpanKindClient), trace.WithAttributes(attribute.String("query", query)))
	defer end(&err)

	tx := readTx(ctx)
	if tx == nil {
		return pgconn.CommandTag{}, ErrNoTransaction
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

func (a App) Bulk(ctx context.Context, fetcher func() ([]any, error), schema, table string, columns ...string) (err error) {
	ctx, end := tracer.StartSpan(ctx, a.tracer, "bulk", trace.WithSpanKind(trace.SpanKindClient), trace.WithAttributes(attribute.String("schema", schema), attribute.String("table", table)))
	defer end(&err)

	tx := readTx(ctx)
	if tx == nil {
		return ErrNoTransaction
	}

	ctx, cancel := context.WithTimeout(ctx, SQLTimeout)
	defer cancel()

	if _, err = tx.CopyFrom(ctx, pgx.Identifier{schema, table}, columns, &feeder{fetcher: fetcher}); err != nil {
		return fmt.Errorf("copy from: %w", err)
	}

	return nil
}
