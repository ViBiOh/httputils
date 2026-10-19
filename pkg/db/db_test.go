package db

import (
	"context"
	"errors"
	"flag"
	"strings"
	"testing"

	"github.com/ViBiOh/httputils/v4/pkg/mocks"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestFlags(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		want string
	}{
		"simple": {
			"Usage of simple:\n  -host string\n    \t[database] Host ${SIMPLE_HOST}\n  -maxConn uint\n    \t[database] Max Open Connections ${SIMPLE_MAX_CONN} (default 20)\n  -minConn uint\n    \t[database] Min Open Connections ${SIMPLE_MIN_CONN} (default 5)\n  -name string\n    \t[database] Name ${SIMPLE_NAME}\n  -pass string\n    \t[database] Pass ${SIMPLE_PASS}\n  -port uint\n    \t[database] Port ${SIMPLE_PORT} (default 5432)\n  -sslmode string\n    \t[database] SSL Mode ${SIMPLE_SSLMODE} (default \"disable\")\n  -user string\n    \t[database] User ${SIMPLE_USER}\n",
		},
	}

	for intention, testCase := range cases {
		t.Run(intention, func(t *testing.T) {
			t.Parallel()

			flagSet := flag.NewFlagSet(intention, flag.ContinueOnError)
			Flags(flagSet, "")

			var writer strings.Builder
			flagSet.SetOutput(&writer)
			flagSet.Usage()

			result := writer.String()

			if result != testCase.want {
				t.Errorf("Flags() = `%s`, want `%s`", result, testCase.want)
			}
		})
	}
}

func TestEnabled(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		instance Service
		want     bool
	}{
		"empty": {
			Service{},
			false,
		},
		"provided": {
			Service{},
			true,
		},
	}

	for intention, testCase := range cases {
		t.Run(intention, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)

			mockDatabase := mocks.NewDatabase(ctrl)

			switch intention {
			case "provided":
				testCase.instance.db = mockDatabase
			}

			if got := testCase.instance.Enabled(); got != testCase.want {
				t.Errorf("Enabled() = %t, want %t", got, testCase.want)
			}
		})
	}
}

func TestPing(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		want bool
	}{
		"simple": {
			true,
		},
		"timeout": {
			false,
		},
	}

	for intention, testCase := range cases {
		t.Run(intention, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)

			mockDatabase := mocks.NewDatabase(ctrl)

			instance := Service{db: mockDatabase}

			switch intention {
			case "simple":
				mockDatabase.EXPECT().Ping(gomock.Any()).Return(nil)
			case "timeout":
				mockDatabase.EXPECT().Ping(gomock.Any()).Return(errors.New("context deadline exceeded"))
			}

			if got := instance.Ping(context.Background()); (got == nil) != testCase.want {
				t.Errorf("Ping() = %t, want %t", got, testCase.want)
			}
		})
	}
}

func TestClose(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		wantErr error
	}{
		"simple": {
			nil,
		},
	}

	for intention := range cases {
		t.Run(intention, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)

			mockDatabase := mocks.NewDatabase(ctrl)

			instance := Service{db: mockDatabase}

			switch intention {
			case "simple":
				mockDatabase.EXPECT().Close()
			}

			instance.Close()
		})
	}
}

func TestReadTx(t *testing.T) {
	t.Parallel()

	var tx pgx.Tx = &pgxpool.Tx{}

	type args struct {
		ctx context.Context
	}

	cases := map[string]struct {
		args args
		want pgx.Tx
	}{
		"empty": {
			args{
				ctx: context.Background(),
			},
			nil,
		},
		"with tx": {
			args{
				ctx: StoreTx(context.Background(), tx),
			},
			tx,
		},
		"not a tx": {
			args{
				ctx: context.WithValue(context.Background(), ctxTxKey, args{}),
			},
			nil,
		},
	}

	for intention, testCase := range cases {
		t.Run(intention, func(t *testing.T) {
			t.Parallel()

			if got := readTx(testCase.args.ctx); got != testCase.want {
				t.Errorf("readTx() = %v, want %v", got, testCase.want)
			}
		})
	}
}

func TestDoAtomic(t *testing.T) {
	t.Parallel()

	type args struct {
		ctx    context.Context
		action func(context.Context) error
	}

	cases := map[string]struct {
		args    args
		wantErr error
	}{
		"no action": {
			args{},
			errors.New("no action provided"),
		},
		"already": {
			args{
				ctx: context.Background(),
				action: func(ctx context.Context) error {
					return nil
				},
			},
			nil,
		},
		"error": {
			args{
				ctx: context.Background(),
				action: func(ctx context.Context) error {
					return nil
				},
			},
			errors.New("no transaction available"),
		},
		"begin": {
			args{
				ctx: context.Background(),
				action: func(ctx context.Context) error {
					return nil
				},
			},
			nil,
		},
		"rollback": {
			args{
				ctx: context.Background(),
				action: func(ctx context.Context) error {
					return errors.New("invalid")
				},
			},
			errors.New("invalid"),
		},
		"rollback error": {
			args{
				ctx: context.Background(),
				action: func(ctx context.Context) error {
					return errors.New("invalid")
				},
			},
			errors.New("invalid\ncannot close transaction"),
		},
	}

	for intention, testCase := range cases {
		t.Run(intention, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)

			mockDatabase := mocks.NewDatabase(ctrl)

			instance := Service{db: mockDatabase}

			ctx := testCase.args.ctx

			switch intention {
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

			gotErr := instance.DoAtomic(ctx, testCase.args.action)

			failed := false

			switch {
			case
				testCase.wantErr == nil && gotErr != nil,
				testCase.wantErr != nil && gotErr == nil,
				testCase.wantErr != nil && gotErr != nil && !strings.Contains(gotErr.Error(), testCase.wantErr.Error()):
				failed = true
			}

			if failed {
				t.Errorf("DoAtomic() = `%s`, want `%s`", gotErr, testCase.wantErr)
			}
		})
	}
}

func TestList(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		wantErr error
	}{
		"simple": {
			nil,
		},
		"error": {
			errors.New("timeout"),
		},
		"read error": {
			errors.New("boom"),
		},
		"tx": {
			nil,
		},
	}

	for intention, testCase := range cases {
		t.Run(intention, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)

			mockDatabase := mocks.NewDatabase(ctrl)

			instance := Service{db: mockDatabase}

			ctx := context.Background()

			switch intention {
			case "simple":
				rows := mocks.NewRows(ctrl)
				rows.EXPECT().Next().Return(true)
				rows.EXPECT().Next().Return(false)
				rows.EXPECT().Err().Return(nil)
				rows.EXPECT().Close()

				mockDatabase.EXPECT().Query(gomock.Any(), "SELECT id FROM item", 1).Return(rows, nil)

			case "error":
				mockDatabase.EXPECT().Query(gomock.Any(), "SELECT id FROM item", 1).Return(nil, errors.New("timeout"))

			case "read error":
				rows := mocks.NewRows(ctrl)
				rows.EXPECT().Next().Return(true)
				rows.EXPECT().Next().Return(false)
				rows.EXPECT().Err().Return(errors.New("boom"))
				rows.EXPECT().Close()

				mockDatabase.EXPECT().Query(gomock.Any(), "SELECT id FROM item", 1).Return(rows, nil)

			case "tx":
				tx := mocks.NewTx(ctrl)
				ctx = StoreTx(ctx, tx)

				rows := mocks.NewRows(ctrl)
				rows.EXPECT().Next().Return(true)
				rows.EXPECT().Next().Return(false)
				rows.EXPECT().Err().Return(nil)
				rows.EXPECT().Close()

				tx.EXPECT().Query(gomock.Any(), "SELECT id FROM item", 1).Return(rows, nil)
			}

			testScanItem := func(row pgx.Rows) error {
				return nil
			}
			gotErr := instance.List(ctx, testScanItem, "SELECT id FROM item", 1)

			failed := false

			switch {
			case
				testCase.wantErr == nil && gotErr != nil,
				testCase.wantErr != nil && gotErr == nil,
				testCase.wantErr != nil && gotErr != nil && !strings.Contains(gotErr.Error(), testCase.wantErr.Error()):
				failed = true
			}

			if failed {
				t.Errorf("Get() = (`%s`), want (`%s`)", gotErr, testCase.wantErr)
			}
		})
	}
}

func TestGet(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		wantErr error
	}{
		"simple": {
			nil,
		},
		"error": {
			errors.New("timeout"),
		},
		"tx": {
			nil,
		},
	}

	for intention, testCase := range cases {
		t.Run(intention, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)

			mockDatabase := mocks.NewDatabase(ctrl)

			instance := Service{db: mockDatabase}

			ctx := context.Background()

			switch intention {
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

			switch {
			case
				testCase.wantErr == nil && gotErr != nil,
				testCase.wantErr != nil && gotErr == nil,
				testCase.wantErr != nil && gotErr != nil && !strings.Contains(gotErr.Error(), testCase.wantErr.Error()):
				failed = true
			}

			if failed {
				t.Errorf("Get() = (`%s`), want (`%s`)", gotErr, testCase.wantErr)
			}
		})
	}
}

func TestCreate(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		wantErr error
	}{
		"no tx": {
			ErrNoTransaction,
		},
		"error": {
			errors.New("timeout"),
		},
		"valid": {
			nil,
		},
	}

	for intention, testCase := range cases {
		t.Run(intention, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)

			mockDatabase := mocks.NewDatabase(ctrl)

			instance := Service{db: mockDatabase}

			ctx := context.Background()

			switch intention {
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

			switch {
			case
				testCase.wantErr == nil && gotErr != nil,
				testCase.wantErr != nil && gotErr == nil,
				testCase.wantErr != nil && gotErr != nil && !strings.Contains(gotErr.Error(), testCase.wantErr.Error()):
				failed = true
			}

			if failed {
				t.Errorf("Create() = (`%s`), want (`%s`)", gotErr, testCase.wantErr)
			}
		})
	}
}

func TestExec(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		wantErr error
	}{
		"error": {
			errors.New("timeout"),
		},
		"valid": {
			nil,
		},
	}

	for intention, testCase := range cases {
		t.Run(intention, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)

			mockDatabase := mocks.NewDatabase(ctrl)

			instance := Service{db: mockDatabase}

			ctx := context.Background()

			switch intention {
			case "error":
				tx := mocks.NewTx(ctrl)
				ctx = StoreTx(ctx, tx)
				tx.EXPECT().Exec(gomock.Any(), "DELETE FROM item WHERE id = $1", 1).Return(pgconn.CommandTag{}, errors.New("timeout"))

			case "valid":
				tx := mocks.NewTx(ctrl)
				ctx = StoreTx(ctx, tx)
				tx.EXPECT().Exec(gomock.Any(), "DELETE FROM item WHERE id = $1", 1).Return(pgconn.CommandTag{}, nil)
			}

			gotErr := instance.Exec(ctx, "DELETE FROM item WHERE id = $1", 1)

			failed := false

			switch {
			case
				testCase.wantErr == nil && gotErr != nil,
				testCase.wantErr != nil && gotErr == nil,
				testCase.wantErr != nil && gotErr != nil && !strings.Contains(gotErr.Error(), testCase.wantErr.Error()):
				failed = true
			}

			if failed {
				t.Errorf("Exec() = `%s`, want `%s`", gotErr, testCase.wantErr)
			}
		})
	}
}

func TestOne(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		wantErr error
	}{
		"error": {
			errors.New("timeout"),
		},
		"zero": {
			errors.New("0 rows affected, wanted 1"),
		},
		"one": {
			nil,
		},
	}

	for intention, testCase := range cases {
		t.Run(intention, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)

			mockDatabase := mocks.NewDatabase(ctrl)

			instance := Service{db: mockDatabase}

			ctx := context.Background()

			switch intention {
			case "error":
				tx := mocks.NewTx(ctrl)
				ctx = StoreTx(ctx, tx)
				tx.EXPECT().Exec(gomock.Any(), "DELETE FROM item WHERE id = $1", 1).Return(pgconn.CommandTag{}, errors.New("timeout"))

			case "zero":
				tx := mocks.NewTx(ctrl)
				ctx = StoreTx(ctx, tx)
				tx.EXPECT().Exec(gomock.Any(), "DELETE FROM item WHERE id = $1", 1).Return(pgconn.CommandTag{}, nil)

			case "one":
				tx := mocks.NewTx(ctrl)
				ctx = StoreTx(ctx, tx)
				tx.EXPECT().Exec(gomock.Any(), "DELETE FROM item WHERE id = $1", 1).Return(pgconn.NewCommandTag("1"), nil)
			}

			gotErr := instance.One(ctx, "DELETE FROM item WHERE id = $1", 1)

			failed := false

			switch {
			case
				testCase.wantErr == nil && gotErr != nil,
				testCase.wantErr != nil && gotErr == nil,
				testCase.wantErr != nil && gotErr != nil && !strings.Contains(gotErr.Error(), testCase.wantErr.Error()):
				failed = true
			}

			if failed {
				t.Errorf("One() = `%s`, want `%s`", gotErr, testCase.wantErr)
			}
		})
	}
}

func TestBulk(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		wantErr error
	}{
		"no tx": {
			ErrNoTransaction,
		},
		"error": {
			errors.New("timeout"),
		},
		"valid": {
			nil,
		},
	}

	for intention, testCase := range cases {
		t.Run(intention, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)

			mockDatabase := mocks.NewDatabase(ctrl)

			instance := Service{db: mockDatabase}

			ctx := context.Background()

			switch intention {
			case "error":
				tx := mocks.NewTx(ctrl)
				ctx = StoreTx(ctx, tx)
				tx.EXPECT().CopyFrom(gomock.Any(), pgx.Identifier{"test", "item"}, []string{"id"}, gomock.Any()).Return(int64(0), errors.New("timeout"))

			case "valid":
				tx := mocks.NewTx(ctrl)
				ctx = StoreTx(ctx, tx)
				tx.EXPECT().CopyFrom(gomock.Any(), pgx.Identifier{"test", "item"}, []string{"id"}, gomock.Any()).Return(int64(0), nil)
			}

			testFeeder := func() ([]any, error) {
				return nil, nil
			}
			gotErr := instance.Bulk(ctx, testFeeder, "test", "item", "id")

			failed := false

			switch {
			case
				testCase.wantErr == nil && gotErr != nil,
				testCase.wantErr != nil && gotErr == nil,
				testCase.wantErr != nil && gotErr != nil && !strings.Contains(gotErr.Error(), testCase.wantErr.Error()):
				failed = true
			}

			if failed {
				t.Errorf("Bulk() = `%s`, want `%s`", gotErr, testCase.wantErr)
			}
		})
	}
}

func TestConnectionString(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		config       Config
		wantHost     string
		wantUser     string
		wantPassword string
		wantDatabase string
		wantPort     uint16
		wantTLS      bool
	}{
		"simple": {
			config:       Config{Host: "db.local", Port: 5432, User: "app", Pass: "secret", Name: "main", SSLMode: "require", MinConn: 2, MaxConn: 5},
			wantHost:     "db.local",
			wantPort:     5432,
			wantUser:     "app",
			wantPassword: "secret",
			wantDatabase: "main",
			wantTLS:      true,
		},
		"special characters": {
			config:       Config{Host: "db.local", Port: 5433, User: "app user", Pass: "p@ss w0rd'&=#%/?\\", Name: "main", SSLMode: "require", MinConn: 2, MaxConn: 5},
			wantHost:     "db.local",
			wantPort:     5433,
			wantUser:     "app user",
			wantPassword: "p@ss w0rd'&=#%/?\\",
			wantDatabase: "main",
			wantTLS:      true,
		},
		"parameter injection in password": {
			config:       Config{Host: "db.local", Port: 5432, User: "app", Pass: "x sslmode=disable host=evil.local", Name: "main", SSLMode: "require", MinConn: 2, MaxConn: 5},
			wantHost:     "db.local",
			wantPort:     5432,
			wantUser:     "app",
			wantPassword: "x sslmode=disable host=evil.local",
			wantDatabase: "main",
			wantTLS:      true,
		},
		"parameter injection in database": {
			config:       Config{Host: "db.local", Port: 5432, User: "app", Pass: "secret", Name: "main?sslmode=disable&host=evil.local", SSLMode: "require", MinConn: 2, MaxConn: 5},
			wantHost:     "db.local",
			wantPort:     5432,
			wantUser:     "app",
			wantPassword: "secret",
			wantDatabase: "main?sslmode=disable&host=evil.local",
			wantTLS:      true,
		},
		"ipv6": {
			config:       Config{Host: "::1", Port: 5432, User: "app", Pass: "secret", Name: "main", SSLMode: "disable", MinConn: 1, MaxConn: 2},
			wantHost:     "::1",
			wantPort:     5432,
			wantUser:     "app",
			wantPassword: "secret",
			wantDatabase: "main",
		},
		"unix socket": {
			config:       Config{Host: "/var/run/postgresql", Port: 5432, User: "app", Pass: "secret", Name: "main", SSLMode: "disable", MinConn: 1, MaxConn: 2},
			wantHost:     "/var/run/postgresql",
			wantPort:     5432,
			wantUser:     "app",
			wantPassword: "secret",
			wantDatabase: "main",
		},
	}

	for intention, testCase := range cases {
		t.Run(intention, func(t *testing.T) {
			t.Parallel()

			parsed, err := pgxpool.ParseConfig(connectionString(&testCase.config))
			require.NoError(t, err)

			assert.Equal(t, testCase.wantHost, parsed.ConnConfig.Host)
			assert.Equal(t, testCase.wantPort, parsed.ConnConfig.Port)
			assert.Equal(t, testCase.wantUser, parsed.ConnConfig.User)
			assert.Equal(t, testCase.wantPassword, parsed.ConnConfig.Password)
			assert.Equal(t, testCase.wantDatabase, parsed.ConnConfig.Database)
			assert.Equal(t, int32(testCase.config.MinConn), parsed.MinConns)
			assert.Equal(t, int32(testCase.config.MaxConn), parsed.MaxConns)
			assert.Equal(t, testCase.wantTLS, parsed.ConnConfig.TLSConfig != nil)
		})
	}
}
