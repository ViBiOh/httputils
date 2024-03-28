package db

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"time"

	"github.com/ViBiOh/flags"
	"github.com/ViBiOh/httputils/v4/pkg/telemetry"
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
	Exec(context.Context, string, ...any) (pgconn.CommandTag, error)
}

type Service struct {
	tracer     trace.Tracer
	db         Database
	attributes []attribute.KeyValue
}

type Config struct {
	Host    string
	User    string
	Pass    string
	Name    string
	SSLMode string
	Port    uint
	MinConn uint
	MaxConn uint
}

func Flags(fs *flag.FlagSet, prefix string, overrides ...flags.Override) *Config {
	var config Config

	flags.New("Host", "Host").Prefix(prefix).DocPrefix("database").StringVar(fs, &config.Host, "", overrides)
	flags.New("Port", "Port").Prefix(prefix).DocPrefix("database").UintVar(fs, &config.Port, 5432, overrides)
	flags.New("User", "User").Prefix(prefix).DocPrefix("database").StringVar(fs, &config.User, "", overrides)
	flags.New("Pass", "Pass").Prefix(prefix).DocPrefix("database").StringVar(fs, &config.Pass, "", overrides)
	flags.New("Name", "Name").Prefix(prefix).DocPrefix("database").StringVar(fs, &config.Name, "", overrides)
	flags.New("MinConn", "Min Open Connections").Prefix(prefix).DocPrefix("database").UintVar(fs, &config.MinConn, 2, overrides)
	flags.New("MaxConn", "Max Open Connections").Prefix(prefix).DocPrefix("database").UintVar(fs, &config.MaxConn, 5, overrides)
	flags.New("Sslmode", "SSL Mode").Prefix(prefix).DocPrefix("database").StringVar(fs, &config.SSLMode, "disable", overrides)

	return &config
}

func New(ctx context.Context, config *Config, tracerProvider trace.TracerProvider) (Service, error) {
	if len(config.Host) == 0 {
		return Service{}, ErrNoHost
	}

	ctx, cancel := context.WithTimeout(ctx, SQLTimeout)
	defer cancel()

	db, err := pgxpool.New(ctx, fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s pool_min_conns=%d pool_max_conns=%d", config.Host, config.Port, config.User, config.Pass, config.Name, config.SSLMode, config.MinConn, config.MaxConn))
	if err != nil {
		return Service{}, fmt.Errorf("connect to postgres: %w", err)
	}

	instance := Service{
		db: db,
	}

	if tracerProvider != nil {
		instance.tracer = tracerProvider.Tracer("database")

		// cf. https://opentelemetry.io/docs/specs/semconv/database/database-spans/
		instance.attributes = []attribute.KeyValue{
			attribute.String("db.system", "postgres"),
			attribute.String("db.name", config.Name),
			attribute.String("server.address", config.Host),
			attribute.Int("server.port", int(config.Port)),
		}
	}

	return instance, instance.Ping(ctx)
}

func (s Service) Enabled() bool {
	return s.db != nil
}

func (s Service) Ping(ctx context.Context) error {
	if !s.Enabled() {
		return nil
	}

	ctx, cancel := context.WithTimeout(ctx, SQLTimeout)
	defer cancel()

	return s.db.Ping(ctx)
}

func (s Service) Close() {
	if !s.Enabled() {
		return
	}

	s.db.Close()
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

func (s Service) DoAtomic(ctx context.Context, action func(context.Context) error) (err error) {
	if action == nil {
		return errors.New("no action provided")
	}

	ctx, end := telemetry.StartSpan(ctx, s.tracer, "transaction", trace.WithSpanKind(trace.SpanKindClient))
	defer end(&err)

	if readTx(ctx) != nil {
		return action(ctx)
	}

	tx, err := s.db.Begin(ctx)
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

func (s Service) Query(ctx context.Context, query string, args ...any) (rows pgx.Rows, err error) {
	ctx, end := telemetry.StartSpan(ctx, s.tracer, "query",
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(
			append([]attribute.KeyValue{attribute.String("db.statement", query)}, s.attributes...)...,
		),
	)
	defer end(&err)

	if tx := readTx(ctx); tx != nil {
		return tx.Query(ctx, query, args...)
	}

	return s.db.Query(ctx, query, args...)
}

func (s Service) List(ctx context.Context, scanner func(pgx.Rows) error, query string, args ...any) error {
	ctx, cancel := context.WithTimeout(ctx, SQLTimeout)
	defer cancel()

	rows, err := s.Query(ctx, query, args...)
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

func (s Service) QueryRow(ctx context.Context, query string, args ...any) pgx.Row {
	ctx, end := telemetry.StartSpan(ctx, s.tracer, "query_row",
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(
			append([]attribute.KeyValue{attribute.String("db.statement", query)}, s.attributes...)...,
		),
	)
	defer end(nil)

	if tx := readTx(ctx); tx != nil {
		return tx.QueryRow(ctx, query, args...)
	}

	return s.db.QueryRow(ctx, query, args...)
}

func (s Service) Get(ctx context.Context, scanner func(pgx.Row) error, query string, args ...any) error {
	ctx, cancel := context.WithTimeout(ctx, SQLTimeout)
	defer cancel()

	return scanner(s.QueryRow(ctx, query, args...))
}

func (s Service) Create(ctx context.Context, query string, args ...any) (id uint64, err error) {
	tx := readTx(ctx)
	if tx == nil {
		return 0, ErrNoTransaction
	}

	ctx, cancel := context.WithTimeout(ctx, SQLTimeout)
	defer cancel()

	var newID uint64

	return newID, tx.QueryRow(ctx, query, args...).Scan(&newID)
}

func (s Service) Exec(ctx context.Context, query string, args ...any) error {
	ctx, cancel := context.WithTimeout(ctx, SQLTimeout)
	defer cancel()

	_, err := s.exec(ctx, query, args...)

	return err
}

func (s Service) One(ctx context.Context, query string, args ...any) error {
	ctx, cancel := context.WithTimeout(ctx, SQLTimeout)
	defer cancel()

	output, err := s.exec(ctx, query, args...)
	if err != nil {
		return err
	}

	if count := output.RowsAffected(); count != 1 {
		return fmt.Errorf("%d rows affected, wanted 1", count)
	}

	return nil
}

func (s Service) exec(ctx context.Context, query string, args ...any) (command pgconn.CommandTag, err error) {
	ctx, end := telemetry.StartSpan(ctx, s.tracer, "exec",
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(
			append([]attribute.KeyValue{attribute.String("db.statement", query)}, s.attributes...)...,
		),
	)
	defer end(&err)

	tx := readTx(ctx)

	if tx == nil {
		return s.db.Exec(ctx, query, args...)
	}

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

func (s Service) Bulk(ctx context.Context, fetcher func() ([]any, error), schema, table string, columns ...string) (err error) {
	ctx, end := telemetry.StartSpan(ctx, s.tracer, "bulk",
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(
			append([]attribute.KeyValue{
				attribute.String("db.schema", schema),
				attribute.String("db.table", table),
			}, s.attributes...)...,
		),
	)
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
