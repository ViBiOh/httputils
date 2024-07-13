package server

import (
	"flag"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFlags(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		want string
	}{
		"simple": {
			"Usage of simple:\n  -address string\n    \t[server] Listen address ${SIMPLE_ADDRESS}\n  -cert string\n    \t[server] Certificate file ${SIMPLE_CERT}\n  -idleTimeout duration\n    \t[server] Idle Timeout ${SIMPLE_IDLE_TIMEOUT} (default 2m0s)\n  -key string\n    \t[server] Key file ${SIMPLE_KEY}\n  -name string\n    \t[server] Name ${SIMPLE_NAME} (default \"http\")\n  -port uint\n    \t[server] Listen port (0 to disable) ${SIMPLE_PORT} (default 1080)\n  -readTimeout duration\n    \t[server] Read Timeout ${SIMPLE_READ_TIMEOUT} (default 5s)\n  -shutdownTimeout duration\n    \t[server] Shutdown Timeout ${SIMPLE_SHUTDOWN_TIMEOUT} (default 10s)\n  -writeTimeout duration\n    \t[server] Write Timeout ${SIMPLE_WRITE_TIMEOUT} (default 10s)\n",
		},
	}

	for intention, testCase := range cases {
		t.Run(intention, func(t *testing.T) {
			t.Parallel()

			fs := flag.NewFlagSet(intention, flag.ContinueOnError)
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

func TestListenAddr(t *testing.T) {
	t.Parallel()

	t.Run("no config", func(t *testing.T) {
		t.Parallel()

		actual := New(&Config{}).ListenAddr()

		assert.Equal(t, "", actual)
	})

	t.Run("minimal config", func(t *testing.T) {
		t.Parallel()

		actual := New(&Config{Port: 1080}).ListenAddr()

		assert.Equal(t, "http://127.0.0.1:1080", actual)
	})

	t.Run("with addr", func(t *testing.T) {
		t.Parallel()

		actual := New(&Config{Port: 80, Address: "localhost"}).ListenAddr()

		assert.Equal(t, "http://localhost:80", actual)
	})

	t.Run("with TLS", func(t *testing.T) {
		t.Parallel()

		actual := New(&Config{Port: 80, Address: "localhost", Cert: "example", Key: "example"}).ListenAddr()

		assert.Equal(t, "https://localhost:80", actual)
	})
}
