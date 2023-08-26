package redis

import (
	"flag"
	"strings"
	"testing"
)

func TestFlags(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		want string
	}{
		"simple": {
			"Usage of simple:\n  -address string slice\n    \t[redis] Redis Address host:port (blank to disable) ${SIMPLE_ADDRESS}, as a string slice, environment variable separated by \",\" (default [127.0.0.1:6379])\n  -database int\n    \t[redis] Redis Database ${SIMPLE_DATABASE}\n  -minIdleConn int\n    \t[redis] Redis Minimum Idle Connections ${SIMPLE_MIN_IDLE_CONN}\n  -password string\n    \t[redis] Redis Password, if any ${SIMPLE_PASSWORD}\n  -poolSize int\n    \t[redis] Redis Pool Size (default GOMAXPROCS*10) ${SIMPLE_POOL_SIZE}\n  -username string\n    \t[redis] Redis Username, if any ${SIMPLE_USERNAME}\n",
		},
	}

	for intention, testCase := range cases {
		intention, testCase := intention, testCase

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
