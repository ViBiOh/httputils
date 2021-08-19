package db

import (
	"context"
	"database/sql"
	"errors"
	"flag"
	"io"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestFlags(t *testing.T) {
	var cases = []struct {
		intention string
		want      string
	}{
		{
			"simple",
			"Usage of simple:\n  -host string\n    \t[database] Host {SIMPLE_HOST}\n  -maxConn uint\n    \t[database] Max Open Connections {SIMPLE_MAX_CONN} (default 5)\n  -name string\n    \t[database] Name {SIMPLE_NAME}\n  -pass string\n    \t[database] Pass {SIMPLE_PASS}\n  -port uint\n    \t[database] Port {SIMPLE_PORT} (default 5432)\n  -sslmode string\n    \t[database] SSL Mode {SIMPLE_SSLMODE} (default \"disable\")\n  -timeout uint\n    \t[database] Connect timeout {SIMPLE_TIMEOUT} (default 10)\n  -user string\n    \t[database] User {SIMPLE_USER}\n",
		},
	}

	for _, tc := range cases {
		t.Run(tc.intention, func(t *testing.T) {
			fs := flag.NewFlagSet(tc.intention, flag.ContinueOnError)
			Flags(fs, "")

			var writer strings.Builder
			fs.SetOutput(&writer)
			fs.Usage()

			result := writer.String()

			if result != tc.want {
				t.Errorf("Flags() = `%s`, want `%s`", result, tc.want)
			}
		})
	}
}

func TestEnabled(t *testing.T) {
	var cases = []struct {
		intention string
		instance  App
		want      bool
	}{
		{
			"empty",
			App{},
			false,
		},
		{
			"provided",
			App{},
			true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.intention, func(t *testing.T) {
			mockDb, _, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
			if err != nil {
				t.Fatalf("unable to create mock database: %s", err)
			}
			defer mockDb.Close()

			if tc.intention == "provided" {
				tc.instance.db = mockDb
			}

			if got := tc.instance.Enabled(); got != tc.want {
				t.Errorf("Enabled() = %t, want %t", got, tc.want)
			}
		})
	}
}

func TestPing(t *testing.T) {
	var cases = []struct {
		intention string
		want      bool
	}{
		{
			"simple",
			true,
		},
		{
			"timeout",
			false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.intention, func(t *testing.T) {
			mockDb, mock, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
			if err != nil {
				t.Fatalf("unable to create mock database: %s", err)
			}
			defer mockDb.Close()

			expectedPing := mock.ExpectPing()

			if tc.intention == "timeout" {
				savedSQLTimeout := SQLTimeout
				SQLTimeout = time.Second
				defer func() {
					SQLTimeout = savedSQLTimeout
				}()

				expectedPing.WillDelayFor(time.Second * 2)
			}

			instance := App{db: mockDb}

			if got := instance.Ping(); (got == nil) != tc.want {
				t.Errorf("Ping() = %t, want %t", got, tc.want)
			}
		})
	}
}

func TestClose(t *testing.T) {
	var cases = []struct {
		intention string
		wantErr   error
	}{
		{
			"simple",
			nil,
		},
		{
			"error",
			errors.New("resource busy"),
		},
	}

	for _, tc := range cases {
		t.Run(tc.intention, func(t *testing.T) {
			mockDb, mock, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
			if err != nil {
				t.Fatalf("unable to create mock database: %s", err)
			}

			expect := mock.ExpectClose()
			if tc.intention == "error" {
				expect.WillReturnError(errors.New("resource busy"))
			}

			gotErr := App{db: mockDb}.Close()

			failed := false

			if tc.wantErr == nil && gotErr != nil {
				failed = true
			} else if tc.wantErr != nil && gotErr == nil {
				failed = true
			} else if tc.wantErr != nil && !strings.Contains(gotErr.Error(), tc.wantErr.Error()) {
				failed = true
			}

			if failed {
				t.Errorf("Close() = `%s`, want `%s`", gotErr, tc.wantErr)
			}
		})
	}
}

func TestReadTx(t *testing.T) {
	mockDb, mock, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
	if err != nil {
		t.Fatalf("unable to create mock database: %s", err)
	}
	defer mockDb.Close()

	mock.ExpectBegin()
	tx, err := mockDb.Begin()
	if err != nil {
		t.Errorf("unable to create tx: %v", err)
	}

	type args struct {
		ctx context.Context
	}

	var cases = []struct {
		intention string
		args      args
		want      *sql.Tx
	}{
		{
			"empty",
			args{
				ctx: context.Background(),
			},
			nil,
		},
		{
			"with tx",
			args{
				ctx: StoreTx(context.Background(), tx),
			},
			tx,
		},
		{
			"not a tx",
			args{
				ctx: context.WithValue(context.Background(), ctxTxKey, mock),
			},
			nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.intention, func(t *testing.T) {
			if got := readTx(tc.args.ctx); got != tc.want {
				t.Errorf("readTx() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestDoAtomic(t *testing.T) {
	type args struct {
		ctx    context.Context
		action func(context.Context) error
	}

	var cases = []struct {
		intention string
		args      args
		wantErr   error
	}{
		{
			"no action",
			args{},
			errors.New("no action provided"),
		},
		{
			"already",
			args{
				ctx: context.Background(),
				action: func(ctx context.Context) error {
					return nil
				},
			},
			nil,
		},
		{
			"error",
			args{
				ctx: context.Background(),
				action: func(ctx context.Context) error {
					return nil
				},
			},
			errors.New("no transaction available"),
		},
		{
			"begin",
			args{
				ctx: context.Background(),
				action: func(ctx context.Context) error {
					return nil
				},
			},
			nil,
		},
		{
			"rollback",
			args{
				ctx: context.Background(),
				action: func(ctx context.Context) error {
					return errors.New("invalid")
				},
			},
			errors.New("invalid"),
		},
		{
			"rollback error",
			args{
				ctx: context.Background(),
				action: func(ctx context.Context) error {
					return errors.New("invalid")
				},
			},
			errors.New("invalid: cannot close transaction"),
		},
	}

	for _, tc := range cases {
		t.Run(tc.intention, func(t *testing.T) {
			mockDb, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("unable to create mock database: %s", err)
			}
			defer mockDb.Close()

			ctx := tc.args.ctx

			if tc.intention == "error" {
				mock.ExpectBegin().WillReturnError(errors.New("no transaction available"))
			} else if tc.intention == "already" {
				mock.ExpectBegin()
				if tx, err := mockDb.Begin(); err != nil {
					t.Errorf("unable to create tx: %v", err)
				} else {
					ctx = StoreTx(ctx, tx)
				}
			} else if tc.intention == "begin" {
				mock.ExpectBegin()
				mock.ExpectCommit()
			} else if tc.intention == "rollback" {
				mock.ExpectBegin()
				mock.ExpectRollback()
			} else if tc.intention == "rollback error" {
				mock.ExpectBegin()
				mock.ExpectRollback().WillReturnError(errors.New("cannot close transaction"))
			}

			instance := App{db: mockDb}
			gotErr := instance.DoAtomic(ctx, tc.args.action)

			failed := false

			if tc.wantErr == nil && gotErr != nil {
				failed = true
			} else if tc.wantErr != nil && gotErr == nil {
				failed = true
			} else if tc.wantErr != nil && gotErr != nil && !strings.Contains(gotErr.Error(), tc.wantErr.Error()) {
				failed = true
			}

			if failed {
				t.Errorf("DoAtomic() = `%s`, want `%s`", gotErr, tc.wantErr)
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("sqlmock unfilled expectations: %s", err)
			}
		})
	}
}

func TestList(t *testing.T) {
	var cases = []struct {
		intention string
		want      []uint64
		wantErr   error
	}{
		{
			"simple",
			[]uint64{1, 2},
			nil,
		},
		{
			"timeout",
			nil,
			sqlmock.ErrCancelled,
		},
		{
			"tx",
			[]uint64{1, 2},
			nil,
		},
		{
			"scan error",
			[]uint64{1, 2},
			errors.New(`converting driver.Value type string ("a") to a uint64: invalid syntax`),
		},
		{
			"close error",
			[]uint64{1, 2},
			errors.New("fetch again"),
		},
	}

	for _, tc := range cases {
		t.Run(tc.intention, func(t *testing.T) {
			mockDb, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("unable to create mock database: %s", err)
			}
			defer mockDb.Close()

			ctx := context.Background()

			if tc.intention == "tx" {
				mock.ExpectBegin()
				if tx, err := mockDb.Begin(); err != nil {
					t.Errorf("unable to create tx: %v", err)
				} else {
					ctx = StoreTx(ctx, tx)
				}
			}

			if tc.intention != "no db" {
				rows := sqlmock.NewRows([]string{"id"}).AddRow(1).AddRow(2)
				expectedQuery := mock.ExpectQuery("SELECT id FROM item").WillReturnRows(rows)

				if tc.intention == "scan error" {
					rows.AddRow("a")
				}

				if tc.intention == "close error" {
					rows.AddRow("b")
					rows.AddRow(3)
					rows.CloseError(errors.New("fetch again"))
				}

				if tc.intention == "timeout" {
					savedSQLTimeout := SQLTimeout
					SQLTimeout = time.Second
					defer func() {
						SQLTimeout = savedSQLTimeout
					}()

					expectedQuery.WillDelayFor(time.Second * 2)
				}
			}

			var got []uint64
			testScanItem := func(row *sql.Rows) error {
				var item uint64
				if err := row.Scan(&item); err != nil {
					return err
				}

				if got == nil {
					got = make([]uint64, 0)
				}
				got = append(got, item)
				return nil
			}

			instance := App{db: mockDb}
			gotErr := instance.List(ctx, testScanItem, "SELECT id FROM item", 1)

			failed := false

			if tc.wantErr == nil && gotErr != nil {
				failed = true
			} else if tc.wantErr != nil && gotErr == nil {
				failed = true
			} else if tc.wantErr != nil && gotErr != nil && !strings.Contains(gotErr.Error(), tc.wantErr.Error()) {
				failed = true
			} else if !reflect.DeepEqual(got, tc.want) {
				failed = true
			}

			if failed {
				t.Errorf("Get() = (%+v, `%s`), want (%+v, `%s`)", got, gotErr, tc.want, tc.wantErr)
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("sqlmock unfilled expectations: %s", err)
			}
		})
	}
}

func TestGet(t *testing.T) {
	var cases = []struct {
		intention string
		want      uint64
		wantErr   error
	}{
		{
			"simple",
			1,
			nil,
		},
		{
			"timeout",
			0,
			sqlmock.ErrCancelled,
		},
		{
			"tx",
			1,
			nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.intention, func(t *testing.T) {
			mockDb, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("unable to create mock database: %s", err)
			}
			defer mockDb.Close()

			ctx := context.Background()

			if tc.intention == "tx" {
				mock.ExpectBegin()
				if tx, err := mockDb.Begin(); err != nil {
					t.Errorf("unable to create tx: %v", err)
				} else {
					ctx = StoreTx(ctx, tx)
				}
			}

			if tc.intention != "no db" {
				expectedQuery := mock.ExpectQuery("SELECT id FROM item WHERE id = ").WithArgs(1).WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

				if tc.intention == "timeout" {
					savedSQLTimeout := SQLTimeout
					SQLTimeout = time.Second
					defer func() {
						SQLTimeout = savedSQLTimeout
					}()

					expectedQuery.WillDelayFor(time.Second * 2)
				}
			}

			var got uint64
			testScanItem := func(row *sql.Row) error {
				return row.Scan(&got)
			}

			instance := App{db: mockDb}
			gotErr := instance.Get(ctx, testScanItem, "SELECT id FROM item WHERE id = $1", 1)

			failed := false

			if tc.wantErr == nil && gotErr != nil {
				failed = true
			} else if tc.wantErr != nil && gotErr == nil {
				failed = true
			} else if tc.wantErr != nil && gotErr != nil && !strings.Contains(gotErr.Error(), tc.wantErr.Error()) {
				failed = true
			} else if got != tc.want {
				failed = true
			}

			if failed {
				t.Errorf("Get() = (%d, `%s`), want (%d, `%s`)", got, gotErr, tc.want, tc.wantErr)
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("sqlmock unfilled expectations: %s", err)
			}
		})
	}
}

func TestCreate(t *testing.T) {

	var cases = []struct {
		intention string
		want      uint64
		wantErr   error
	}{
		{
			"no tx",
			0,
			ErrNoTransaction,
		},
		{
			"timeout",
			0,
			sqlmock.ErrCancelled,
		},
		{
			"valid",
			1,
			nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.intention, func(t *testing.T) {
			mockDb, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("unable to create mock database: %s", err)
			}
			defer mockDb.Close()

			ctx := context.Background()

			if tc.intention != "no tx" {
				mock.ExpectBegin()
				if tx, err := mockDb.Begin(); err != nil {
					t.Errorf("unable to create tx: %v", err)
				} else {
					ctx = StoreTx(ctx, tx)
				}

				expectedQuery := mock.ExpectQuery("INSERT INTO item VALUES").WithArgs(1).WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

				if tc.intention == "timeout" {
					savedSQLTimeout := SQLTimeout
					SQLTimeout = time.Second
					defer func() {
						SQLTimeout = savedSQLTimeout
					}()

					expectedQuery.WillDelayFor(time.Second * 2)
				}
			}

			instance := App{db: mockDb}
			got, gotErr := instance.Create(ctx, "INSERT INTO item VALUES ($1)", 1)

			failed := false

			if tc.wantErr == nil && gotErr != nil {
				failed = true
			} else if tc.wantErr != nil && gotErr == nil {
				failed = true
			} else if tc.wantErr != nil && gotErr != nil && !strings.Contains(gotErr.Error(), tc.wantErr.Error()) {
				failed = true
			} else if got != tc.want {
				failed = true
			}

			if failed {
				t.Errorf("Create() = (%d, `%s`), want (%d, `%s`)", got, gotErr, tc.want, tc.wantErr)
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("sqlmock unfilled expectations: %s", err)
			}
		})
	}
}

func TestExec(t *testing.T) {
	var cases = []struct {
		intention string
		wantErr   error
	}{
		{
			"no tx",
			ErrNoTransaction,
		},
		{
			"timeout",
			sqlmock.ErrCancelled,
		},
		{
			"valid",
			nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.intention, func(t *testing.T) {
			mockDb, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("unable to create mock database: %s", err)
			}
			defer mockDb.Close()

			ctx := context.Background()

			if tc.intention != "no tx" {
				mock.ExpectBegin()
				if tx, err := mockDb.Begin(); err != nil {
					t.Errorf("unable to create tx: %v", err)
				} else {
					ctx = StoreTx(ctx, tx)
				}

				expectedQuery := mock.ExpectExec("DELETE FROM item WHERE id = (.+)").WithArgs(1).WillReturnResult(sqlmock.NewResult(0, 1))

				if tc.intention == "timeout" {
					savedSQLTimeout := SQLTimeout
					SQLTimeout = time.Second
					defer func() {
						SQLTimeout = savedSQLTimeout
					}()

					expectedQuery.WillDelayFor(time.Second * 2)
				}
			}

			instance := App{db: mockDb}
			gotErr := instance.Exec(ctx, "DELETE FROM item WHERE id = $1", 1)

			failed := false

			if tc.wantErr == nil && gotErr != nil {
				failed = true
			} else if tc.wantErr != nil && gotErr == nil {
				failed = true
			} else if tc.wantErr != nil && gotErr != nil && !strings.Contains(gotErr.Error(), tc.wantErr.Error()) {
				failed = true
			}

			if failed {
				t.Errorf("Exec() = `%s`, want `%s`", gotErr, tc.wantErr)
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("sqlmock unfilled expectations: %s", err)
			}
		})
	}
}

func TestBulk(t *testing.T) {
	type args struct {
		feeder  func(*sql.Stmt) error
		schema  string
		table   string
		columns []string
	}

	var count uint

	var cases = []struct {
		intention string
		args      args
		wantErr   error
	}{
		{
			"no tx",
			args{},
			ErrNoTransaction,
		},
		{
			"prepare error",
			args{
				schema:  "business",
				table:   "users",
				columns: []string{"name", "email"},
			},
			errors.New("unable to prepare context: invalid statement"),
		},
		{
			"feed error",
			args{
				feeder: func(*sql.Stmt) error {
					return errors.New("unknown error")
				},
				schema:  "business",
				table:   "users",
				columns: []string{"name", "email"},
			},
			errors.New("unable to feed bulk creation: unknown error"),
		},
		{
			"exec error",
			args{
				feeder: func(stmt *sql.Stmt) error {
					if count == 0 {
						_, err := stmt.Exec("vibioh", "nobody@localhost")
						count++
						return err
					}
					return ErrBulkEnded
				},
				schema:  "business",
				table:   "users",
				columns: []string{"name", "email"},
			},
			errors.New("unable to exec bulk creation: invalid values"),
		},
		{
			"close error",
			args{
				feeder: func(stmt *sql.Stmt) error {
					if count == 0 {
						_, err := stmt.Exec("vibioh", "nobody@localhost")
						count++
						return err
					}
					return ErrBulkEnded
				},
				schema:  "business",
				table:   "users",
				columns: []string{"name", "email"},
			},
			errors.New("error while closing"),
		},
		{
			"success",
			args{
				feeder: func(stmt *sql.Stmt) error {
					if count == 0 {
						_, err := stmt.Exec("vibioh", "nobody@localhost")
						count++
						return err
					}
					return ErrBulkEnded
				},
				schema:  "business",
				table:   "users",
				columns: []string{"name", "email"},
			},
			nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.intention, func(t *testing.T) {
			mockDb, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("unable to create mock database: %s", err)
			}
			defer mockDb.Close()

			ctx := context.Background()

			if tc.intention != "no tx" {
				mock.ExpectBegin()
				if tx, err := mockDb.Begin(); err != nil {
					t.Errorf("unable to create tx: %v", err)
				} else {
					ctx = StoreTx(ctx, tx)
				}

				prepareStmt := mock.ExpectPrepare(`COPY "business"\."users" \("name", "email"\) FROM STDIN`)
				if tc.intention == "prepare error" {
					prepareStmt.WillReturnError(errors.New("invalid statement"))
				}

				if tc.intention == "exec error" || tc.intention == "close error" || tc.intention == "success" {
					prepareStmt.ExpectExec().WithArgs("vibioh", "nobody@localhost").WillReturnResult(sqlmock.NewResult(0, 1))

					exec := prepareStmt.ExpectExec()
					if tc.intention == "exec error" {
						exec.WillReturnError(errors.New("invalid values"))
					} else {
						exec.WillReturnResult(sqlmock.NewResult(0, 1))

						if tc.intention == "close error" {
							prepareStmt.WillReturnCloseError(errors.New("error while closing"))
						}
					}
				}
			}

			count = 0
			instance := App{db: mockDb}
			gotErr := instance.Bulk(ctx, tc.args.feeder, tc.args.schema, tc.args.table, tc.args.columns...)

			failed := false

			if tc.wantErr == nil && gotErr != nil {
				failed = true
			} else if tc.wantErr != nil && gotErr == nil {
				failed = true
			} else if tc.wantErr != nil && gotErr != nil && !strings.Contains(gotErr.Error(), tc.wantErr.Error()) {
				failed = true
			}

			if failed {
				t.Errorf("Bulk() = `%s`, want `%s`", gotErr, tc.wantErr)
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("sqlmock unfilled expectations: %s", err)
			}
		})
	}
}

type errClose func() error

func (ec errClose) Close() error {
	return errors.New("close error")
}

func TestSafeClose(t *testing.T) {
	type args struct {
		closer io.Closer
		err    error
	}

	var cases = []struct {
		intention string
		args      args
		wantErr   error
	}{
		{
			"no error",
			args{
				closer: io.NopCloser(strings.NewReader("")),
				err:    nil,
			},
			nil,
		},
		{
			"close error",
			args{
				closer: new(errClose),
				err:    nil,
			},
			errors.New("close error"),
		},
		{
			"nested error",
			args{
				closer: new(errClose),
				err:    sql.ErrNoRows,
			},
			sql.ErrNoRows,
		},
	}

	for _, tc := range cases {
		t.Run(tc.intention, func(t *testing.T) {
			gotErr := safeClose(tc.args.closer, tc.args.err)

			failed := false

			if tc.wantErr == nil && gotErr != nil {
				failed = true
			} else if tc.wantErr != nil && gotErr == nil {
				failed = true
			} else if tc.wantErr != nil && errors.Is(gotErr, tc.wantErr) {
				failed = true
			}

			if failed {
				t.Errorf("safeClose() = (`%s`), want (`%s`)", gotErr, tc.wantErr)
			}
		})
	}
}
