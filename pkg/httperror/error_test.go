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
	t.Parallel()

	cases := map[string]struct {
		err        error
		want       string
		wantStatus int
	}{
		"should set body and status": {
			errors.New("bad request"),
			"bad request\n",
			http.StatusBadRequest,
		},
	}

	for intention, testCase := range cases {
		intention := intention
		testCase := testCase

		t.Run(intention, func(t *testing.T) {
			t.Parallel()

			writer := httptest.NewRecorder()
			BadRequest(writer, testCase.err)

			if result := writer.Code; result != testCase.wantStatus {
				t.Errorf("BadRequest() = %d, want status %d", result, testCase.wantStatus)
			}

			if result, _ := request.ReadBodyResponse(writer.Result()); string(result) != testCase.want {
				t.Errorf("BadRequest() = `%s`, want `%s`", string(result), testCase.want)
			}
		})
	}
}

func TestUnauthorized(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		err        error
		want       string
		wantStatus int
	}{
		"should set body and status": {
			errors.New("unauthorized"),
			"unauthorized\n",
			http.StatusUnauthorized,
		},
	}

	for intention, testCase := range cases {
		intention := intention
		testCase := testCase

		t.Run(intention, func(t *testing.T) {
			t.Parallel()

			writer := httptest.NewRecorder()
			Unauthorized(writer, testCase.err)

			if result := writer.Code; result != testCase.wantStatus {
				t.Errorf("Unauthorized() = %d, want status %d", result, testCase.wantStatus)
			}

			if result, _ := request.ReadBodyResponse(writer.Result()); string(result) != testCase.want {
				t.Errorf("Unauthorized() = `%s`, want `%s`", string(result), testCase.want)
			}
		})
	}
}

func TestForbidden(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		want       string
		wantStatus int
	}{
		"should set body and status": {
			"⛔️\n",
			http.StatusForbidden,
		},
	}

	for intention, testCase := range cases {
		intention := intention
		testCase := testCase

		t.Run(intention, func(t *testing.T) {
			t.Parallel()

			writer := httptest.NewRecorder()
			Forbidden(writer)

			if result := writer.Code; result != testCase.wantStatus {
				t.Errorf("Forbidden() = %d, want status %d", result, testCase.wantStatus)
			}

			if result, _ := request.ReadBodyResponse(writer.Result()); string(result) != testCase.want {
				t.Errorf("Forbidden() = `%s`, want `%s`", string(result), testCase.want)
			}
		})
	}
}

func TestNotFound(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		want       string
		wantStatus int
	}{
		"should set body and status": {
			"¯\\_(ツ)_/¯\n",
			http.StatusNotFound,
		},
	}

	for intention, testCase := range cases {
		intention := intention
		testCase := testCase

		t.Run(intention, func(t *testing.T) {
			t.Parallel()

			writer := httptest.NewRecorder()
			NotFound(writer)

			if result := writer.Code; result != testCase.wantStatus {
				t.Errorf("NotFound() = %d, want status %d", result, testCase.wantStatus)
			}

			if result, _ := request.ReadBodyResponse(writer.Result()); string(result) != testCase.want {
				t.Errorf("NotFound() = `%s`, want `%s`", string(result), testCase.want)
			}
		})
	}
}

func TestInternalServerError(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		err        error
		want       string
		wantStatus int
	}{
		"should set body and status": {
			errors.New("failed to do something"),
			"Oops! Something went wrong. Server's logs contain more details.\n",
			http.StatusInternalServerError,
		},
	}

	for intention, testCase := range cases {
		intention := intention
		testCase := testCase

		t.Run(intention, func(t *testing.T) {
			t.Parallel()

			writer := httptest.NewRecorder()
			InternalServerError(writer, testCase.err)

			if result := writer.Code; result != testCase.wantStatus {
				t.Errorf("InternalServerError() = %d, want status %d", result, testCase.wantStatus)
			}

			if result, _ := request.ReadBodyResponse(writer.Result()); string(result) != testCase.want {
				t.Errorf("InternalServerError() = `%s`, want `%s`", string(result), testCase.want)
			}
		})
	}
}

func TestHandleError(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		err         error
		want        bool
		wantStatus  int
		wantMessage string
	}{
		"no error": {
			nil,
			false,
			http.StatusOK,
			"",
		},
		"invalid valud": {
			model.WrapInvalid(errors.New("invalid value")),
			true,
			http.StatusBadRequest,
			"invalid value: invalid\n",
		},
		"invalid auth": {
			model.WrapUnauthorized(errors.New("invalid auth")),
			true,
			http.StatusUnauthorized,
			"invalid auth: unauthorized\n",
		},
		"invalid creds": {
			model.WrapForbidden(errors.New("invalid credentials")),
			true,
			http.StatusForbidden,
			"invalid credentials: forbidden\n",
		},
		"not found": {
			model.WrapNotFound(errors.New("unknown id")),
			true,
			http.StatusNotFound,
			"¯\\_(ツ)_/¯\n",
		},
		"method not allowed": {
			model.WrapMethodNotAllowed(errors.New("unknown method")),
			true,
			http.StatusMethodNotAllowed,
			"unknown method: method not allowed\n",
		},
		"internal server error": {
			errors.New("bool"),
			true,
			http.StatusInternalServerError,
			"Oops! Something went wrong. Server's logs contain more details.\n",
		},
	}

	for intention, testCase := range cases {
		intention := intention
		testCase := testCase

		t.Run(intention, func(t *testing.T) {
			t.Parallel()

			writer := httptest.NewRecorder()

			if got := HandleError(writer, testCase.err); got != testCase.want {
				t.Errorf("HandleError = %t, want %t", got, testCase.want)
			}

			if got := writer.Code; got != testCase.wantStatus {
				t.Errorf("HandleError = HTTP/%d, want HTTP/%d", got, testCase.wantStatus)
			}

			if got, _ := request.ReadBodyResponse(writer.Result()); string(got) != testCase.wantMessage {
				t.Errorf("HandleError = `%s`, want `%s`", string(got), testCase.wantMessage)
			}
		})
	}
}

func TestErrorStatus(t *testing.T) {
	t.Parallel()

	type args struct {
		err error
	}

	cases := map[string]struct {
		args        args
		want        int
		wantMessage string
	}{
		"nil": {
			args{
				err: nil,
			},
			http.StatusInternalServerError,
			"",
		},
		"simple": {
			args{
				err: errors.New("boom"),
			},
			http.StatusInternalServerError,
			internalError,
		},
		"invalid": {
			args{
				err: model.WrapInvalid(errors.New("bad request")),
			},
			http.StatusBadRequest,
			"bad request",
		},
		"unauthorized": {
			args{
				err: model.WrapUnauthorized(errors.New("jwt missing")),
			},
			http.StatusUnauthorized,
			"unauthorized",
		},
		"forbidden": {
			args{
				err: model.WrapForbidden(errors.New("jwt invalid")),
			},
			http.StatusForbidden,
			"forbidden",
		},
		"not found": {
			args{
				err: model.WrapNotFound(errors.New("unknown")),
			},
			http.StatusNotFound,
			"unknown",
		},
		"method": {
			args{
				err: model.WrapMethodNotAllowed(errors.New("not allowed")),
			},
			http.StatusMethodNotAllowed,
			"not allowed",
		},
		"internal": {
			args{
				err: model.WrapInternal(errors.New("boom")),
			},
			http.StatusMethodNotAllowed,
			internalError,
		},
	}

	for intention, testCase := range cases {
		intention := intention
		testCase := testCase

		t.Run(intention, func(t *testing.T) {
			t.Parallel()

			if got, gotMessage := ErrorStatus(testCase.args.err); got != testCase.want && gotMessage != testCase.wantMessage {
				t.Errorf("ErrorStatus() = (%d, `%s`), want (%d, `%s`)", got, gotMessage, testCase.want, testCase.wantMessage)
			}
		})
	}
}
