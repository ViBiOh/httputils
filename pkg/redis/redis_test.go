package redis

import (
	"flag"
	"strings"
	"testing"
)

func TestFlags(t *testing.T) {
	cases := []struct {
		intention string
		want      string
	}{
		{
			"simple",
			"Usage of simple:\n  -address string\n    \t[redis] Redis Address {SIMPLE_ADDRESS} (default \"localhost:6379\")\n  -alias string\n    \t[redis] Connection alias, for metric {SIMPLE_ALIAS}\n  -database int\n    \t[redis] Redis Database {SIMPLE_DATABASE}\n  -password string\n    \t[redis] Redis Password, if any {SIMPLE_PASSWORD}\n",
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
