package db

import (
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
	}

	for _, tc := range cases {
		t.Run(tc.intention, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("unable to create mock database: %s", err)
			}
			defer db.Close()

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
			gotErr := GetRow(db, "SELECT id FROM item WHERE id = $1", 1).Scan(&got)

			failed := false

			if tc.wantErr != nil && !errors.Is(gotErr, tc.wantErr) {
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
		tx bool
	}

	var cases = []struct {
		intention string
		args      args
		want      uint64
		wantErr   error
	}{
		{
			"simple",
			args{
				tx: false,
			},
			1,
			nil,
		},
		{
			"timeout",
			args{
				tx: false,
			},
			0,
			sqlmock.ErrCancelled,
		},
		{
			"with tx",
			args{
				tx: true,
			},
			1,
			nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.intention, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("unable to create mock database: %s", err)
			}
			defer db.Close()

			var tx *sql.Tx
			if tc.args.tx {
				mock.ExpectBegin()
				dbTx, err := db.Begin()

				if err != nil {
					t.Errorf("unable to create tx: %v", err)
				}
				tx = dbTx
			}

			if !tc.args.tx {
				mock.ExpectBegin()
			}

			expectedQuery := mock.ExpectQuery("INSERT INTO item VALUES").WithArgs(1).WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

			if !tc.args.tx {
				if tc.wantErr != nil {
					mock.ExpectRollback()
				} else {
					mock.ExpectCommit()
				}
			}

			if tc.intention == "timeout" {
				savedSQLTimeout := SQLTimeout
				SQLTimeout = time.Second
				defer func() {
					SQLTimeout = savedSQLTimeout
				}()

				expectedQuery.WillDelayFor(time.Second * 2)
			}

			got, gotErr := Create(db, tx, "INSERT INTO item VALUES ($1)", 1)

			failed := false

			if tc.wantErr != nil && !errors.Is(gotErr, tc.wantErr) {
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
		tx bool
	}

	var cases = []struct {
		intention string
		args      args
		wantErr   error
	}{
		{
			"simple",
			args{
				tx: false,
			},
			nil,
		},
		{
			"timeout",
			args{
				tx: false,
			},
			sqlmock.ErrCancelled,
		},
		{
			"with tx",
			args{
				tx: true,
			},
			nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.intention, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("unable to create mock database: %s", err)
			}
			defer db.Close()

			var tx *sql.Tx
			if tc.args.tx {
				mock.ExpectBegin()
				dbTx, err := db.Begin()

				if err != nil {
					t.Errorf("unable to create tx: %v", err)
				}
				tx = dbTx
			}

			if !tc.args.tx {
				mock.ExpectBegin()
			}

			expectedQuery := mock.ExpectExec("DELETE FROM item WHERE id = (.+)").WithArgs(1).WillReturnResult(sqlmock.NewResult(0, 1))

			if !tc.args.tx {
				if tc.wantErr != nil {
					mock.ExpectRollback()
				} else {
					mock.ExpectCommit()
				}
			}

			if tc.intention == "timeout" {
				savedSQLTimeout := SQLTimeout
				SQLTimeout = time.Second
				defer func() {
					SQLTimeout = savedSQLTimeout
				}()

				expectedQuery.WillDelayFor(time.Second * 2)
			}

			gotErr := Exec(db, tx, "DELETE FROM item WHERE id = $1", 1)

			failed := false

			if tc.wantErr != nil && !errors.Is(gotErr, tc.wantErr) {
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
