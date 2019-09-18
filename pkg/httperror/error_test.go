package httperror

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ViBiOh/httputils/v2/pkg/request"
)

func TestBadRequest(t *testing.T) {
	var cases = []struct {
		intention  string
		err        error
		want       string
		wantStatus int
	}{
		{
			"should set body and status",
			errors.New("bad request"),
			"bad request\n",
			http.StatusBadRequest,
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.intention, func(t *testing.T) {
			writer := httptest.NewRecorder()
			BadRequest(writer, testCase.err)

			if result := writer.Code; result != testCase.wantStatus {
				t.Errorf("BadRequest(%#v) = %#v, want status %#v", testCase.err, result, testCase.wantStatus)
			}

			if result, _ := request.ReadBodyResponse(writer.Result()); string(result) != testCase.want {
				t.Errorf("BadRequest(%#v) = %#v, want %#v", testCase.err, string(result), testCase.want)
			}
		})
	}
}

func TestUnauthorized(t *testing.T) {
	var cases = []struct {
		intention  string
		err        error
		want       string
		wantStatus int
	}{
		{
			"should set body and status",
			errors.New("unauthorized"),
			"unauthorized\n",
			http.StatusUnauthorized,
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.intention, func(t *testing.T) {
			writer := httptest.NewRecorder()
			Unauthorized(writer, testCase.err)

			if result := writer.Code; result != testCase.wantStatus {
				t.Errorf("Unauthorized(%#v) = %#v, want status %#v", testCase.err, result, testCase.wantStatus)
			}

			if result, _ := request.ReadBodyResponse(writer.Result()); string(result) != testCase.want {
				t.Errorf("Unauthorized(%#v) = %#v, want %#v", testCase.err, string(result), testCase.want)
			}
		})
	}
}

func TestForbidden(t *testing.T) {
	var cases = []struct {
		intention  string
		want       string
		wantStatus int
	}{
		{
			"should set body and status",
			"⛔️\n",
			http.StatusForbidden,
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.intention, func(t *testing.T) {
			writer := httptest.NewRecorder()
			Forbidden(writer)

			if result := writer.Code; result != testCase.wantStatus {
				t.Errorf("Forbidden() = %#v, want status %#v", result, testCase.wantStatus)
			}

			if result, _ := request.ReadBodyResponse(writer.Result()); string(result) != testCase.want {
				t.Errorf("Forbidden() = %#v, want %#v", string(result), testCase.want)
			}
		})
	}
}

func TestNotFound(t *testing.T) {
	var cases = []struct {
		intention  string
		want       string
		wantStatus int
	}{
		{
			"should set body and status",
			"¯\\_(ツ)_/¯\n",
			http.StatusNotFound,
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.intention, func(t *testing.T) {
			writer := httptest.NewRecorder()
			NotFound(writer)

			if result := writer.Code; result != testCase.wantStatus {
				t.Errorf("NotFound() = %#v, want status %#v", result, testCase.wantStatus)
			}

			if result, _ := request.ReadBodyResponse(writer.Result()); string(result) != testCase.want {
				t.Errorf("NotFound() = %#v, want %#v", string(result), testCase.want)
			}
		})
	}
}

func TestInternalServerError(t *testing.T) {
	var cases = []struct {
		intention  string
		err        error
		want       string
		wantStatus int
	}{
		{
			"should set body and status",
			errors.New("failed to do something"),
			"failed to do something\n",
			http.StatusInternalServerError,
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.intention, func(t *testing.T) {
			writer := httptest.NewRecorder()
			InternalServerError(writer, testCase.err)

			if result := writer.Code; result != testCase.wantStatus {
				t.Errorf("InternalServerError(%#v) = %#v, want status %#v", testCase.err, result, testCase.wantStatus)
			}

			if result, _ := request.ReadBodyResponse(writer.Result()); string(result) != testCase.want {
				t.Errorf("InternalServerError(%#v) = %#v, want %#v", testCase.err, string(result), testCase.want)
			}
		})
	}
}
