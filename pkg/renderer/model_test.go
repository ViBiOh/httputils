package renderer

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestParseMessage(t *testing.T) {
	type args struct {
		r *http.Request
	}

	var cases = []struct {
		intention string
		args      args
		want      Message
	}{
		{
			"empty",
			args{
				r: httptest.NewRequest(http.MethodGet, fmt.Sprintf("/?%s", NewSuccessMessage("")), nil),
			},
			Message{},
		},
		{
			"success",
			args{
				r: httptest.NewRequest(http.MethodGet, fmt.Sprintf("/?%s", NewSuccessMessage("HelloWorld")), nil),
			},
			NewSuccessMessage("HelloWorld"),
		},
		{
			"error",
			args{
				r: httptest.NewRequest(http.MethodGet, fmt.Sprintf("/?%s", NewErrorMessage("HelloWorld")), nil),
			},
			NewErrorMessage("HelloWorld"),
		},
	}

	for _, tc := range cases {
		t.Run(tc.intention, func(t *testing.T) {
			if got := ParseMessage(tc.args.r); got != tc.want {
				t.Errorf("ParseMessage() = %v, want %v", got, tc.want)
			}
		})
	}
}
