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
			`Usage of simple:
  -address string slice
    	[redis] Redis Address host:port (blank to disable) ${SIMPLE_ADDRESS}, as a string slice, environment variable separated by "," (default [127.0.0.1:6379])
  -database int
    	[redis] Redis Database ${SIMPLE_DATABASE}
  -password string
    	[redis] Redis Password, if any ${SIMPLE_PASSWORD}
  -username string
    	[redis] Redis Username, if any ${SIMPLE_USERNAME}
`,
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
