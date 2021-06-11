package db

import (
	"context"
	"database/sql"
	"errors"
	"flag"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/ViBiOh/httputils/v4/pkg/flags"
	"github.com/lib/pq"
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

// App of package
type App interface {
	Ping() error
	Close() error
	DoAtomic(ctx context.Context, action func(context.Context) error) error
	List(ctx context.Context, scanner func(*sql.Rows) error, query string, args ...interface{}) error
	Get(ctx context.Context, scanner func(*sql.Row) error, query string, args ...interface{}) error
	Create(ctx context.Context, query string, args ...interface{}) (uint64, error)
	Exec(ctx context.Context, query string, args ...interface{}) error
	Bulk(ctx context.Context, feeder func(*sql.Stmt) error, schema, table string, columns ...string) error
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
}

type app struct {
	db *sql.DB
}

// Flags adds flags for configuring package
func Flags(fs *flag.FlagSet, prefix string, overrides ...flags.Override) Config {
	return Config{
		host:    flags.New(prefix, "database").Name("Host").Default(flags.Default("Host", "", overrides)).Label("Host").ToString(fs),
		port:    flags.New(prefix, "database").Name("Port").Default(flags.Default("Port", 5432, overrides)).Label("Port").ToUint(fs),
		user:    flags.New(prefix, "database").Name("User").Default(flags.Default("User", "", overrides)).Label("User").ToString(fs),
		pass:    flags.New(prefix, "database").Name("Pass").Default(flags.Default("Pass", "", overrides)).Label("Pass").ToString(fs),
		name:    flags.New(prefix, "database").Name("Name").Default(flags.Default("Name", "", overrides)).Label("Name").ToString(fs),
		maxConn: flags.New(prefix, "database").Name("MaxConn").Default(flags.Default("MaxConn", 5, overrides)).Label("Max Open Connections").ToUint(fs),
		sslmode: flags.New(prefix, "database").Name("Sslmode").Default(flags.Default("Sslmode", "disable", overrides)).Label("SSL Mode").ToString(fs),
	}
}

// New creates new App from Config
func New(config Config) (App, error) {
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
	db.SetMaxOpenConns(int(*config.maxConn))

	instance := app{
		db: db,
	}

	return instance, instance.Ping()
}

// Ping indicate if database is ready or not
func (a app) Ping() error {
	ctx, cancel := context.WithTimeout(context.Background(), SQLTimeout)
	defer cancel()

	return a.db.PingContext(ctx)
}

// Close the database connection
func (a app) Close() error {
	return a.db.Close()
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
func (a app) DoAtomic(ctx context.Context, action func(context.Context) error) (err error) {
	if action == nil {
		return errors.New("no action provided")
	}

	if readTx(ctx) != nil {
		return action(ctx)
	}

	var tx *sql.Tx

	tx, err = a.db.Begin()
	if err != nil {
		return
	}

	defer func() {
		if err == nil {
			err = tx.Commit()
		} else if rollbackErr := tx.Rollback(); rollbackErr != nil {
			err = fmt.Errorf("%s: %w", err.Error(), rollbackErr)
		}
	}()

	err = action(StoreTx(ctx, tx))
	return
}

// List execute multiple rows query
func (a app) List(ctx context.Context, scanner func(*sql.Rows) error, query string, args ...interface{}) (err error) {
	ctx, cancel := context.WithTimeout(ctx, SQLTimeout)
	defer cancel()

	var rows *sql.Rows

	if tx := readTx(ctx); tx != nil {
		rows, err = tx.QueryContext(ctx, query, args...)
	} else {
		rows, err = a.db.QueryContext(ctx, query, args...)
	}

	if err != nil {
		return err
	}

	defer func() {
		err = safeClose(rows, err)
	}()

	for rows.Next() && err == nil {
		err = scanner(rows)
	}

	return
}

// Get execute single row query
func (a app) Get(ctx context.Context, scanner func(*sql.Row) error, query string, args ...interface{}) error {
	ctx, cancel := context.WithTimeout(ctx, SQLTimeout)
	defer cancel()

	if tx := readTx(ctx); tx != nil {
		return scanner(tx.QueryRowContext(ctx, query, args...))
	}

	return scanner(a.db.QueryRowContext(ctx, query, args...))
}

// Create execute query with a RETURNING id
func (a app) Create(ctx context.Context, query string, args ...interface{}) (uint64, error) {
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
func (a app) Exec(ctx context.Context, query string, args ...interface{}) error {
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
func (a app) Bulk(ctx context.Context, feeder func(*sql.Stmt) error, schema, table string, columns ...string) (err error) {
	tx := readTx(ctx)
	if tx == nil {
		return ErrNoTransaction
	}

	var stmt *sql.Stmt

	stmt, err = tx.Prepare(pq.CopyInSchema(schema, table, columns...))
	if err != nil {
		return fmt.Errorf("unable to prepare context: %s", err)
	}

	defer func() {
		err = safeClose(stmt, err)
	}()

	for err == nil {
		err = feeder(stmt)
	}

	if err == ErrBulkEnded {
		err = nil
	} else {
		err = fmt.Errorf("unable to feed bulk creation: %s", err)
		return
	}

	ctx, cancel := context.WithTimeout(ctx, SQLTimeout)
	defer cancel()

	if _, err = stmt.ExecContext(ctx); err != nil {
		err = fmt.Errorf("unable to exec bulk creation: %s", err)
	}

	return
}

func safeClose(closer io.Closer, err error) error {
	if closeErr := closer.Close(); closeErr != nil {
		if err == nil {
			return closeErr
		}

		return fmt.Errorf("%s: %w", err, closeErr)
	}

	return err
}
