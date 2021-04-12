package httperror

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ViBiOh/httputils/v4/pkg/model"
	"github.com/ViBiOh/httputils/v4/pkg/request"
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

	for _, tc := range cases {
		t.Run(tc.intention, func(t *testing.T) {
			writer := httptest.NewRecorder()
			BadRequest(writer, tc.err)

			if result := writer.Code; result != tc.wantStatus {
				t.Errorf("BadRequest() = %d, want status %d", result, tc.wantStatus)
			}

			if result, _ := request.ReadBodyResponse(writer.Result()); string(result) != tc.want {
				t.Errorf("BadRequest() = `%s`, want `%s`", string(result), tc.want)
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

	for _, tc := range cases {
		t.Run(tc.intention, func(t *testing.T) {
			writer := httptest.NewRecorder()
			Unauthorized(writer, tc.err)

			if result := writer.Code; result != tc.wantStatus {
				t.Errorf("Unauthorized() = %d, want status %d", result, tc.wantStatus)
			}

			if result, _ := request.ReadBodyResponse(writer.Result()); string(result) != tc.want {
				t.Errorf("Unauthorized() = `%s`, want `%s`", string(result), tc.want)
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

	for _, tc := range cases {
		t.Run(tc.intention, func(t *testing.T) {
			writer := httptest.NewRecorder()
			Forbidden(writer)

			if result := writer.Code; result != tc.wantStatus {
				t.Errorf("Forbidden() = %d, want status %d", result, tc.wantStatus)
			}

			if result, _ := request.ReadBodyResponse(writer.Result()); string(result) != tc.want {
				t.Errorf("Forbidden() = `%s`, want `%s`", string(result), tc.want)
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

	for _, tc := range cases {
		t.Run(tc.intention, func(t *testing.T) {
			writer := httptest.NewRecorder()
			NotFound(writer)

			if result := writer.Code; result != tc.wantStatus {
				t.Errorf("NotFound() = %d, want status %d", result, tc.wantStatus)
			}

			if result, _ := request.ReadBodyResponse(writer.Result()); string(result) != tc.want {
				t.Errorf("NotFound() = `%s`, want `%s`", string(result), tc.want)
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
			"Oops! Something went wrong. Server's logs contain more details.\n",
			http.StatusInternalServerError,
		},
	}

	for _, tc := range cases {
		t.Run(tc.intention, func(t *testing.T) {
			writer := httptest.NewRecorder()
			InternalServerError(writer, tc.err)

			if result := writer.Code; result != tc.wantStatus {
				t.Errorf("InternalServerError() = %d, want status %d", result, tc.wantStatus)
			}

			if result, _ := request.ReadBodyResponse(writer.Result()); string(result) != tc.want {
				t.Errorf("InternalServerError() = `%s`, want `%s`", string(result), tc.want)
			}
		})
	}
}

func TestErrorStatus(t *testing.T) {
	type args struct {
		err error
	}

	var cases = []struct {
		intention   string
		args        args
		want        int
		wantMessage string
	}{
		{
			"nil",
			args{
				err: nil,
			},
			http.StatusInternalServerError,
			"",
		},
		{
			"simple",
			args{
				err: errors.New("boom"),
			},
			http.StatusInternalServerError,
			internalError,
		},
		{
			"invalid",
			args{
				err: model.WrapInvalid(errors.New("bad request")),
			},
			http.StatusBadRequest,
			"bad request",
		},
		{
			"unauthorized",
			args{
				err: model.WrapUnauthorized(errors.New("jwt missing")),
			},
			http.StatusUnauthorized,
			"unauthorized",
		},
		{
			"forbidden",
			args{
				err: model.WrapForbidden(errors.New("jwt invalid")),
			},
			http.StatusForbidden,
			"forbidden",
		},
		{
			"not found",
			args{
				err: model.WrapNotFound(errors.New("unknown")),
			},
			http.StatusNotFound,
			"unknown",
		},
		{
			"method",
			args{
				err: model.WrapMethodNotAllowed(errors.New("not allowed")),
			},
			http.StatusMethodNotAllowed,
			"not allowed",
		},
		{
			"internal",
			args{
				err: model.WrapInternal(errors.New("boom")),
			},
			http.StatusMethodNotAllowed,
			internalError,
		},
	}

	for _, tc := range cases {
		t.Run(tc.intention, func(t *testing.T) {
			if got, gotMessage := ErrorStatus(tc.args.err); got != tc.want && gotMessage != tc.wantMessage {
				t.Errorf("ErrorStatus() = (%d, `%s`), want (%d, `%s`)", got, gotMessage, tc.want, tc.wantMessage)
			}
		})
	}
}
