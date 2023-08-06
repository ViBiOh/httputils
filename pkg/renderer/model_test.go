package renderer

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestEtag(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		instance Page
		want     string
	}{
		"simple": {
			NewPage("index", http.StatusOK, nil),
			"a0317bc82594800f",
		},
	}

	for intention, testCase := range cases {
		intention, testCase := intention, testCase

		t.Run(intention, func(t *testing.T) {
			t.Parallel()

			if got := testCase.instance.etag(); got != testCase.want {
				t.Errorf("Etag() = `%s`, want `%s`", got, testCase.want)
			}
		})
	}
}

func TestParseMessage(t *testing.T) {
	t.Parallel()

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

	for intention, testCase := range cases {
		intention, testCase := intention, testCase

		t.Run(intention, func(t *testing.T) {
			t.Parallel()

			if got := ParseMessage(testCase.args.r); got != testCase.want {
				t.Errorf("ParseMessage() = %v, want %v", got, testCase.want)
			}
		})
	}
}

type testStruct struct {
	ID    uint64
	Name  string
	Items []string
}

func (ts testStruct) String() string {
	return ""
}

func BenchmarkEtag(b *testing.B) {
	page := Page{
		Content: map[string]any{
			"Version": "localhost",
			"Date":    "Today",
			"Hash":    "deadbeef",
			"Items": []testStruct{
				{ID: 8000, Name: "John", Items: []string{"one", "two", "three"}},
			},
		},
		Template: "index",
		Status:   http.StatusOK,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		page.etag()
	}
}
