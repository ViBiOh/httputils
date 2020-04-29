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
			if got := ReadTx(tc.args.ctx); got != tc.want {
				t.Errorf("ReadTx() = %v, want %v", got, tc.want)
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

			expectedQuery := mock.ExpectQuery("SELECT id FROM item WHERE id = ").WithArgs(1).WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

			if tc.intention == "timeout" {
				savedSQLTimeout := SQLTimeout
				SQLTimeout = time.Second
				defer func() {
					SQLTimeout = savedSQLTimeout
				}()

				expectedQuery.WillDelayFor(time.Second * 2)
			}

			var got uint64
			testScanItem := func(row RowScanner) error {
				return row.Scan(&got)
			}
			gotErr := GetRow(ctx, mockDb, testScanItem, "SELECT id FROM item WHERE id = $1", 1)

			failed := false

			if tc.wantErr == nil && gotErr != nil {
				failed = true
			} else if tc.wantErr != nil && !errors.Is(gotErr, tc.wantErr) {
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
	type args struct {
		addTxOnContext  bool
		errorOnCreateTx bool
	}

	var cases = []struct {
		intention string
		args      args
		want      uint64
		wantErr   error
	}{
		{
			"simple",
			args{},
			1,
			nil,
		},
		{
			"timeout",
			args{},
			0,
			sqlmock.ErrCancelled,
		},
		{
			"with tx",
			args{
				addTxOnContext: true,
			},
			1,
			nil,
		},
		{
			"with tx error",
			args{
				errorOnCreateTx: true,
			},
			0,
			errors.New("call to database transaction Begin was not expected"),
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

			if tc.args.addTxOnContext {
				mock.ExpectBegin()
				if tx, err := mockDb.Begin(); err != nil {
					t.Errorf("unable to create tx: %v", err)
				} else {
					ctx = StoreTx(ctx, tx)
				}
			} else if !tc.args.errorOnCreateTx {
				mock.ExpectBegin()
			}

			if !tc.args.errorOnCreateTx {
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

			if !tc.args.addTxOnContext && !tc.args.errorOnCreateTx {
				if tc.wantErr != nil {
					mock.ExpectRollback()
				} else {
					mock.ExpectCommit()
				}
			}

			got, gotErr := Create(ctx, mockDb, "INSERT INTO item VALUES ($1)", 1)

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
	type args struct {
		addTxOnContext  bool
		errorOnCreateTx bool
	}

	var cases = []struct {
		intention string
		args      args
		wantErr   error
	}{
		{
			"simple",
			args{},
			nil,
		},
		{
			"timeout",
			args{},
			sqlmock.ErrCancelled,
		},
		{
			"with tx",
			args{
				addTxOnContext: true,
			},
			nil,
		},
		{
			"with tx error",
			args{
				errorOnCreateTx: true,
			},
			errors.New("call to database transaction Begin was not expected"),
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

			if tc.args.addTxOnContext {
				mock.ExpectBegin()
				if tx, err := mockDb.Begin(); err != nil {
					t.Errorf("unable to create tx: %v", err)
				} else {
					ctx = StoreTx(ctx, tx)
				}
			} else if !tc.args.errorOnCreateTx {
				mock.ExpectBegin()
			}

			if !tc.args.errorOnCreateTx {
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

			if !tc.args.addTxOnContext && !tc.args.errorOnCreateTx {
				if tc.wantErr != nil {
					mock.ExpectRollback()
				} else {
					mock.ExpectCommit()
				}
			}

			gotErr := Exec(ctx, mockDb, "DELETE FROM item WHERE id = $1", 1)

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
