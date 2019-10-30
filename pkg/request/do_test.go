package request

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestDoAndRead(t *testing.T) {
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/simple" {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("valid"))
			return
		}

		if r.URL.Path == "/invalid" {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("invalid"))
			return
		}

		if r.URL.Path == "/internalError" {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}))
	defer testServer.Close()

	simple, _ := http.NewRequest(http.MethodGet, testServer.URL+"/simple", nil)
	invalid, _ := http.NewRequest(http.MethodGet, testServer.URL+"/invalid", nil)
	internalError, _ := http.NewRequest(http.MethodGet, testServer.URL+"/internalError", nil)

	var cases = []struct {
		intention  string
		ctx        context.Context
		request    *http.Request
		want       string
		wantStatus int
		wantErr    error
	}{
		{
			"simple",
			context.Background(),
			simple,
			"valid",
			http.StatusOK,
			nil,
		},
		{
			"invalid",
			context.Background(),
			invalid,
			"",
			http.StatusBadRequest,
			errors.New("HTTP/400\ninvalid"),
		},
		{
			"internalError",
			context.Background(),
			internalError,
			"",
			http.StatusInternalServerError,
			errors.New("HTTP/500"),
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.intention, func(t *testing.T) {
			rawResult, status, _, err := DoAndRead(testCase.ctx, testCase.request)

			result, _ := ReadBody(rawResult)

			failed := false

			if testCase.wantErr != nil && (err == nil || err.Error() != testCase.wantErr.Error()) {
				failed = true
			} else if string(result) != testCase.want {
				failed = true
			} else if status != testCase.wantStatus {
				failed = true
			}

			if failed {
				t.Errorf("DoAndRead() = (%s, %d, %#v), want (%s, %d, %#v)", result, status, err, testCase.want, testCase.wantStatus, testCase.wantErr)
			}
		})
	}
}
