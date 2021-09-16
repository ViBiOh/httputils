package db

import (
	"context"
	"errors"
	"flag"
	"strings"
	"testing"

	"github.com/ViBiOh/httputils/v4/pkg/mocks"
	"github.com/golang/mock/gomock"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
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
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockDatabase := mocks.NewDatabase(ctrl)

			switch tc.intention {
			case "provided":
				tc.instance.db = mockDatabase
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
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockDatabase := mocks.NewDatabase(ctrl)

			instance := App{db: mockDatabase}

			switch tc.intention {
			case "simple":
				mockDatabase.EXPECT().Ping(gomock.Any()).Return(nil)
			case "timeout":
				mockDatabase.EXPECT().Ping(gomock.Any()).Return(errors.New("context deadline exceeded"))
			}

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
	}

	for _, tc := range cases {
		t.Run(tc.intention, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockDatabase := mocks.NewDatabase(ctrl)

			instance := App{db: mockDatabase}

			switch tc.intention {
			case "simple":
				mockDatabase.EXPECT().Close()
			}

			instance.Close()
		})
	}
}

func TestReadTx(t *testing.T) {
	var tx pgx.Tx = &pgxpool.Tx{}

	type args struct {
		ctx context.Context
	}

	var cases = []struct {
		intention string
		args      args
		want      pgx.Tx
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
				ctx: context.WithValue(context.Background(), ctxTxKey, args{}),
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
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockDatabase := mocks.NewDatabase(ctrl)

			instance := App{db: mockDatabase}

			ctx := tc.args.ctx

			switch tc.intention {
			case "error":
				mockDatabase.EXPECT().Begin(gomock.Any()).Return(nil, errors.New("no transaction available"))
			case "already":
				tx := mocks.NewTx(ctrl)
				ctx = StoreTx(ctx, tx)
			case "begin":
				tx := mocks.NewTx(ctrl)
				mockDatabase.EXPECT().Begin(gomock.Any()).Return(tx, nil)
				tx.EXPECT().Commit(gomock.Any()).Return(nil)
			case "rollback":
				tx := mocks.NewTx(ctrl)
				mockDatabase.EXPECT().Begin(gomock.Any()).Return(tx, nil)
				tx.EXPECT().Rollback(gomock.Any()).Return(nil)
			case "rollback error":
				tx := mocks.NewTx(ctrl)
				mockDatabase.EXPECT().Begin(gomock.Any()).Return(tx, nil)
				tx.EXPECT().Rollback(gomock.Any()).Return(errors.New("cannot close transaction"))
			}

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
		})
	}
}

func TestList(t *testing.T) {
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
			errors.New("timeout"),
		},
		{
			"tx",
			nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.intention, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockDatabase := mocks.NewDatabase(ctrl)

			instance := App{db: mockDatabase}

			ctx := context.Background()

			switch tc.intention {
			case "simple":
				rows := mocks.NewRows(ctrl)
				rows.EXPECT().Next().Return(true)
				rows.EXPECT().Next().Return(false)
				rows.EXPECT().Close()

				mockDatabase.EXPECT().Query(gomock.Any(), "SELECT id FROM item", 1).Return(rows, nil)

			case "error":
				mockDatabase.EXPECT().Query(gomock.Any(), "SELECT id FROM item", 1).Return(nil, errors.New("timeout"))

			case "tx":
				tx := mocks.NewTx(ctrl)
				ctx = StoreTx(ctx, tx)

				rows := mocks.NewRows(ctrl)
				rows.EXPECT().Next().Return(true)
				rows.EXPECT().Next().Return(false)
				rows.EXPECT().Close()

				tx.EXPECT().Query(gomock.Any(), "SELECT id FROM item", 1).Return(rows, nil)
			}

			testScanItem := func(row pgx.Rows) error {
				return nil
			}
			gotErr := instance.List(ctx, testScanItem, "SELECT id FROM item", 1)

			failed := false

			if tc.wantErr == nil && gotErr != nil {
				failed = true
			} else if tc.wantErr != nil && gotErr == nil {
				failed = true
			} else if tc.wantErr != nil && gotErr != nil && !strings.Contains(gotErr.Error(), tc.wantErr.Error()) {
				failed = true
			}

			if failed {
				t.Errorf("Get() = (`%s`), want (`%s`)", gotErr, tc.wantErr)
			}
		})
	}
}

func TestGet(t *testing.T) {
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
			errors.New("timeout"),
		},
		{
			"tx",
			nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.intention, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockDatabase := mocks.NewDatabase(ctrl)

			instance := App{db: mockDatabase}

			ctx := context.Background()

			switch tc.intention {
			case "simple":
				row := mocks.NewRow(ctrl)
				row.EXPECT().Scan().Return(nil)

				mockDatabase.EXPECT().QueryRow(gomock.Any(), "SELECT id FROM item WHERE id = $1", 1).Return(row)

			case "error":
				row := mocks.NewRow(ctrl)
				row.EXPECT().Scan().Return(errors.New("timeout"))
				mockDatabase.EXPECT().QueryRow(gomock.Any(), "SELECT id FROM item WHERE id = $1", 1).Return(row)

			case "tx":
				tx := mocks.NewTx(ctrl)
				ctx = StoreTx(ctx, tx)

				row := mocks.NewRow(ctrl)
				row.EXPECT().Scan().Return(nil)

				tx.EXPECT().QueryRow(gomock.Any(), "SELECT id FROM item WHERE id = $1", 1).Return(row)
			}

			testScanItem := func(row pgx.Row) error {
				return row.Scan()
			}
			gotErr := instance.Get(ctx, testScanItem, "SELECT id FROM item WHERE id = $1", 1)

			failed := false

			if tc.wantErr == nil && gotErr != nil {
				failed = true
			} else if tc.wantErr != nil && gotErr == nil {
				failed = true
			} else if tc.wantErr != nil && gotErr != nil && !strings.Contains(gotErr.Error(), tc.wantErr.Error()) {
				failed = true
			}

			if failed {
				t.Errorf("Get() = (`%s`), want (`%s`)", gotErr, tc.wantErr)
			}
		})
	}
}

func TestCreate(t *testing.T) {
	var cases = []struct {
		intention string
		wantErr   error
	}{
		{
			"no tx",
			ErrNoTransaction,
		},
		{
			"error",
			errors.New("timeout"),
		},
		{
			"valid",
			nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.intention, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockDatabase := mocks.NewDatabase(ctrl)

			instance := App{db: mockDatabase}

			ctx := context.Background()

			switch tc.intention {
			case "error":
				tx := mocks.NewTx(ctrl)
				ctx = StoreTx(ctx, tx)

				row := mocks.NewRow(ctrl)
				row.EXPECT().Scan(gomock.Any()).Return(errors.New("timeout"))
				tx.EXPECT().QueryRow(gomock.Any(), "INSERT INTO item VALUES ($1)", 1).Return(row)

			case "valid":
				tx := mocks.NewTx(ctrl)
				ctx = StoreTx(ctx, tx)

				row := mocks.NewRow(ctrl)
				row.EXPECT().Scan(gomock.Any()).Return(nil)

				tx.EXPECT().QueryRow(gomock.Any(), "INSERT INTO item VALUES ($1)", 1).Return(row)
			}

			_, gotErr := instance.Create(ctx, "INSERT INTO item VALUES ($1)", 1)

			failed := false

			if tc.wantErr == nil && gotErr != nil {
				failed = true
			} else if tc.wantErr != nil && gotErr == nil {
				failed = true
			} else if tc.wantErr != nil && gotErr != nil && !strings.Contains(gotErr.Error(), tc.wantErr.Error()) {
				failed = true
			}

			if failed {
				t.Errorf("Create() = (`%s`), want (`%s`)", gotErr, tc.wantErr)
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
			"error",
			errors.New("timeout"),
		},
		{
			"valid",
			nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.intention, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockDatabase := mocks.NewDatabase(ctrl)

			instance := App{db: mockDatabase}

			ctx := context.Background()

			switch tc.intention {
			case "error":
				tx := mocks.NewTx(ctrl)
				ctx = StoreTx(ctx, tx)
				tx.EXPECT().Exec(gomock.Any(), "DELETE FROM item WHERE id = $1", 1).Return(nil, errors.New("timeout"))

			case "valid":
				tx := mocks.NewTx(ctrl)
				ctx = StoreTx(ctx, tx)
				tx.EXPECT().Exec(gomock.Any(), "DELETE FROM item WHERE id = $1", 1).Return(nil, nil)
			}

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
		})
	}
}

func TestBulk(t *testing.T) {

	var cases = []struct {
		intention string
		wantErr   error
	}{
		{
			"no tx",
			ErrNoTransaction,
		},
		{
			"error",
			errors.New("timeout"),
		},
		{
			"valid",
			nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.intention, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockDatabase := mocks.NewDatabase(ctrl)

			instance := App{db: mockDatabase}

			ctx := context.Background()

			switch tc.intention {
			case "error":
				tx := mocks.NewTx(ctrl)
				ctx = StoreTx(ctx, tx)
				tx.EXPECT().CopyFrom(gomock.Any(), pgx.Identifier{"test", "item"}, []string{"id"}, gomock.Any()).Return(int64(0), errors.New("timeout"))

			case "valid":
				tx := mocks.NewTx(ctrl)
				ctx = StoreTx(ctx, tx)
				tx.EXPECT().CopyFrom(gomock.Any(), pgx.Identifier{"test", "item"}, []string{"id"}, gomock.Any()).Return(int64(0), nil)
			}

			testFeeder := func() ([]interface{}, error) {
				return nil, nil
			}
			gotErr := instance.Bulk(ctx, testFeeder, "test", "item", "id")

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
		})
	}
}
