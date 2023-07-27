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
			"Usage of simple:\n  -address string\n    \t[redis] Redis Address host:port (blank to disable) ${SIMPLE_ADDRESS} (default \"localhost:6379\")\n  -alias string\n    \t[redis] Connection alias, for metric ${SIMPLE_ALIAS}\n  -database int\n    \t[redis] Redis Database ${SIMPLE_DATABASE}\n  -password string\n    \t[redis] Redis Password, if any ${SIMPLE_PASSWORD}\n  -pipelineSize int\n    \t[redis] Redis Pipeline Size ${SIMPLE_PIPELINE_SIZE} (default 50)\n  -username string\n    \t[redis] Redis Username, if any ${SIMPLE_USERNAME}\n",
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
