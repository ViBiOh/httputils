package db

import (
	"context"
	"database/sql"
	"errors"
	"flag"
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
			"Usage of simple:\n  -host string\n    \t[database] Host {SIMPLE_HOST}\n  -name string\n    \t[database] Name {SIMPLE_NAME}\n  -pass string\n    \t[database] Pass {SIMPLE_PASS}\n  -port uint\n    \t[database] Port {SIMPLE_PORT} (default 5432)\n  -sslmode string\n    \t[database] SSL Mode {SIMPLE_SSLMODE} (default \"disable\")\n  -user string\n    \t[database] User {SIMPLE_USER}\n",
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.intention, func(t *testing.T) {
			fs := flag.NewFlagSet(testCase.intention, flag.ContinueOnError)
			Flags(fs, "")

			var writer strings.Builder
			fs.SetOutput(&writer)
			fs.Usage()

			result := writer.String()

			if result != testCase.want {
				t.Errorf("Flags() = `%s`, want `%s`", result, testCase.want)
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
			db, mock, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
			if err != nil {
				t.Fatalf("unable to create mock database: %s", err)
			}
			defer db.Close()

			expectedPing := mock.ExpectPing()

			if tc.intention == "timeout" {
				savedSQLTimeout := SQLTimeout
				SQLTimeout = time.Second
				defer func() {
					SQLTimeout = savedSQLTimeout
				}()

				expectedPing.WillDelayFor(time.Second * 2)
			}

			if got := Ping(db); got != tc.want {
				t.Errorf("Ping() = %t, want %t", got, tc.want)
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
		db     *sql.DB
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

			gotErr := DoAtomic(ctx, mockDb, tc.args.action)

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

func TestGetRow(t *testing.T) {
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
		{
			"no db",
			0,
			errors.New("no transaction or database provided"),
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

			usedDb := mockDb
			if tc.intention == "no db" {
				usedDb = nil
			}

			var got uint64
			testScanItem := func(row RowScanner) error {
				return row.Scan(&got)
			}
			gotErr := GetRow(ctx, usedDb, testScanItem, "SELECT id FROM item WHERE id = $1", 1)

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
				t.Errorf("GetRow() = (%d, `%s`), want (%d, `%s`)", got, gotErr, tc.want, tc.wantErr)
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
			errors.New("no transaction in context, please wrap with DoAtomic()"),
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

			got, gotErr := Create(ctx, "INSERT INTO item VALUES ($1)", 1)

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
			errors.New("no transaction in context, please wrap with DoAtomic()"),
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

			gotErr := Exec(ctx, "DELETE FROM item WHERE id = $1", 1)

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
