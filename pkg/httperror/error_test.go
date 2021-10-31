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
	cases := []struct {
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
	cases := []struct {
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
	cases := []struct {
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
	cases := []struct {
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
	cases := []struct {
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

func TestHandleError(t *testing.T) {
	cases := []struct {
		intention   string
		err         error
		want        bool
		wantStatus  int
		wantMessage string
	}{
		{
			"no error",
			nil,
			false,
			http.StatusOK,
			"",
		},
		{
			"invalid",
			model.WrapInvalid(errors.New("invalid value")),
			true,
			http.StatusBadRequest,
			"invalid value: invalid\n",
		},
		{
			"invalid",
			model.WrapUnauthorized(errors.New("invalid auth")),
			true,
			http.StatusUnauthorized,
			"invalid auth: unauthorized\n",
		},
		{
			"invalid",
			model.WrapForbidden(errors.New("invalid credentials")),
			true,
			http.StatusForbidden,
			"invalid credentials: forbidden\n",
		},
		{
			"not found",
			model.WrapNotFound(errors.New("unknown id")),
			true,
			http.StatusNotFound,
			"¯\\_(ツ)_/¯\n",
		},
		{
			"method not allowed",
			model.WrapMethodNotAllowed(errors.New("unknown method")),
			true,
			http.StatusMethodNotAllowed,
			"unknown method: method not allowed\n",
		},
		{
			"internal server error",
			errors.New("bool"),
			true,
			http.StatusInternalServerError,
			"Oops! Something went wrong. Server's logs contain more details.\n",
		},
	}

	for _, tc := range cases {
		t.Run(tc.intention, func(t *testing.T) {
			writer := httptest.NewRecorder()

			if got := HandleError(writer, tc.err); got != tc.want {
				t.Errorf("HandleError = %t, want %t", got, tc.want)
			}

			if got := writer.Code; got != tc.wantStatus {
				t.Errorf("HandleError = HTTP/%d, want HTTP/%d", got, tc.wantStatus)
			}

			if got, _ := request.ReadBodyResponse(writer.Result()); string(got) != tc.wantMessage {
				t.Errorf("HandleError = `%s`, want `%s`", string(got), tc.wantMessage)
			}
		})
	}
}

func TestErrorStatus(t *testing.T) {
	type args struct {
		err error
	}

	cases := []struct {
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
