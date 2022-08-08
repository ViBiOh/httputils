package renderer

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestEtag(t *testing.T) {
	cases := map[string]struct {
		instance Page
		want     string
	}{
		"simple": {
			NewPage("index", http.StatusOK, nil),
			"b1c86216f372f99944c9bbcd9cc99cd6224556b506e275d8be6b03f0316bbbfd",
		},
	}

	for intention, tc := range cases {
		t.Run(intention, func(t *testing.T) {
			if got := tc.instance.etag(); got != tc.want {
				t.Errorf("Etag() = `%s`, want `%s`", got, tc.want)
			}
		})
	}
}

func TestParseMessage(t *testing.T) {
	type args struct {
		r *http.Request
	}

	cases := map[string]struct {
		args args
		want Message
	}{
		"empty": {
			args{
				r: httptest.NewRequest(http.MethodGet, fmt.Sprintf("/?%s", NewSuccessMessage("")), nil),
			},
			Message{},
		},
		"success": {
			args{
				r: httptest.NewRequest(http.MethodGet, fmt.Sprintf("/?%s", NewSuccessMessage("HelloWorld")), nil),
			},
			NewSuccessMessage("HelloWorld"),
		},
		"error": {
			args{
				r: httptest.NewRequest(http.MethodGet, fmt.Sprintf("/?%s", NewErrorMessage("HelloWorld")), nil),
			},
			NewErrorMessage("HelloWorld"),
		},
	}

	for intention, tc := range cases {
		t.Run(intention, func(t *testing.T) {
			if got := ParseMessage(tc.args.r); got != tc.want {
				t.Errorf("ParseMessage() = %v, want %v", got, tc.want)
			}
		})
	}
}
