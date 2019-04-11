package httperror

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ViBiOh/httputils/pkg/request"
)

func Test_BadRequest(t *testing.T) {
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
		writer := httptest.NewRecorder()
		BadRequest(writer, testCase.err)

		if result := writer.Code; result != testCase.wantStatus {
			t.Errorf("%+v\nBadRequest(%+v) = %+v, want status %+v", testCase.intention, testCase.err, result, testCase.wantStatus)
		}

		if result, _ := request.ReadBodyResponse(writer.Result()); string(result) != testCase.want {
			t.Errorf("%+v\nBadRequest(%+v) = %+v, want %+v", testCase.intention, testCase.err, string(result), testCase.want)
		}
	}
}

func Test_Unauthorized(t *testing.T) {
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
		writer := httptest.NewRecorder()
		Unauthorized(writer, testCase.err)

		if result := writer.Code; result != testCase.wantStatus {
			t.Errorf("%+v\nUnauthorized(%+v) = %+v, want status %+v", testCase.intention, testCase.err, result, testCase.wantStatus)
		}

		if result, _ := request.ReadBodyResponse(writer.Result()); string(result) != testCase.want {
			t.Errorf("%+v\nUnauthorized(%+v) = %+v, want %+v", testCase.intention, testCase.err, string(result), testCase.want)
		}
	}
}

func Test_Forbidden(t *testing.T) {
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
		writer := httptest.NewRecorder()
		Forbidden(writer)

		if result := writer.Code; result != testCase.wantStatus {
			t.Errorf("%+v\nForbidden() = %+v, want status %+v", testCase.intention, result, testCase.wantStatus)
		}

		if result, _ := request.ReadBodyResponse(writer.Result()); string(result) != testCase.want {
			t.Errorf("%+v\nForbidden() = %+v, want %+v", testCase.intention, string(result), testCase.want)
		}
	}
}

func Test_NotFound(t *testing.T) {
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
		writer := httptest.NewRecorder()
		NotFound(writer)

		if result := writer.Code; result != testCase.wantStatus {
			t.Errorf("%+v\nNotFound() = %+v, want status %+v", testCase.intention, result, testCase.wantStatus)
		}

		if result, _ := request.ReadBodyResponse(writer.Result()); string(result) != testCase.want {
			t.Errorf("%+v\nNotFound() = %+v, want %+v", testCase.intention, string(result), testCase.want)
		}
	}
}

func Test_InternalServerError(t *testing.T) {
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
		writer := httptest.NewRecorder()
		InternalServerError(writer, testCase.err)

		if result := writer.Code; result != testCase.wantStatus {
			t.Errorf("%+v\nInternalServerError(%+v) = %+v, want status %+v", testCase.intention, testCase.err, result, testCase.wantStatus)
		}

		if result, _ := request.ReadBodyResponse(writer.Result()); string(result) != testCase.want {
			t.Errorf("%+v\nInternalServerError(%+v) = %+v, want %+v", testCase.intention, testCase.err, string(result), testCase.want)
		}
	}
}
