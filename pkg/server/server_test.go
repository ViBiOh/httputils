package server

import (
	"flag"
	"strings"
	"testing"
)

func TestFlags(t *testing.T) {
	cases := map[string]struct {
		want string
	}{
		"simple": {
			"Usage of simple:\n  -address string\n    \t[server] Listen address {SIMPLE_ADDRESS}\n  -cert string\n    \t[server] Certificate file {SIMPLE_CERT}\n  -idleTimeout string\n    \t[server] Idle Timeout {SIMPLE_IDLE_TIMEOUT} (default \"2m\")\n  -key string\n    \t[server] Key file {SIMPLE_KEY}\n  -port uint\n    \t[server] Listen port (0 to disable) {SIMPLE_PORT} (default 1080)\n  -readTimeout string\n    \t[server] Read Timeout {SIMPLE_READ_TIMEOUT} (default \"5s\")\n  -shutdownTimeout string\n    \t[server] Shutdown Timeout {SIMPLE_SHUTDOWN_TIMEOUT} (default \"10s\")\n  -writeTimeout string\n    \t[server] Write Timeout {SIMPLE_WRITE_TIMEOUT} (default \"10s\")\n",
		},
	}

	for intention, tc := range cases {
		t.Run(intention, func(t *testing.T) {
			fs := flag.NewFlagSet(intention, flag.ContinueOnError)
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
