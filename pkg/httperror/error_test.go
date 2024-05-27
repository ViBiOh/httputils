package httperror

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
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
		t.Run(intention, func(t *testing.T) {
			t.Parallel()

			writer := httptest.NewRecorder()
			BadRequest(context.Background(), writer, testCase.err)

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
		"should set default message": {
			nil,
			"🙅\n",
			http.StatusUnauthorized,
		},
	}

	for intention, testCase := range cases {
		t.Run(intention, func(t *testing.T) {
			t.Parallel()

			writer := httptest.NewRecorder()
			Unauthorized(context.Background(), writer, testCase.err)

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
		t.Run(intention, func(t *testing.T) {
			t.Parallel()

			writer := httptest.NewRecorder()
			Forbidden(context.Background(), writer)

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
			"🤷\n",
			http.StatusNotFound,
		},
	}

	for intention, testCase := range cases {
		t.Run(intention, func(t *testing.T) {
			t.Parallel()

			writer := httptest.NewRecorder()
			NotFound(context.Background(), writer)

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
		t.Run(intention, func(t *testing.T) {
			t.Parallel()

			writer := httptest.NewRecorder()
			InternalServerError(context.Background(), writer, testCase.err)

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
			"invalid value\ninvalid\n",
		},
		"invalid auth": {
			model.WrapUnauthorized(errors.New("invalid auth")),
			true,
			http.StatusUnauthorized,
			"invalid auth\nunauthorized\n",
		},
		"invalid creds": {
			model.WrapForbidden(errors.New("invalid credentials")),
			true,
			http.StatusForbidden,
			"invalid credentials\nforbidden\n",
		},
		"not found": {
			model.WrapNotFound(errors.New("unknown id")),
			true,
			http.StatusNotFound,
			"🤷\n",
		},
		"method not allowed": {
			model.WrapMethodNotAllowed(errors.New("unknown method")),
			true,
			http.StatusMethodNotAllowed,
			"unknown method\nmethod not allowed\n",
		},
		"internal server error": {
			errors.New("bool"),
			true,
			http.StatusInternalServerError,
			"Oops! Something went wrong. Server's logs contain more details.\n",
		},
	}

	for intention, testCase := range cases {
		t.Run(intention, func(t *testing.T) {
			t.Parallel()

			writer := httptest.NewRecorder()

			if got := HandleError(context.Background(), writer, testCase.err); got != testCase.want {
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
		t.Run(intention, func(t *testing.T) {
			t.Parallel()

			if got, gotMessage := ErrorStatus(testCase.args.err); got != testCase.want && gotMessage != testCase.wantMessage {
				t.Errorf("ErrorStatus() = (%d, `%s`), want (%d, `%s`)", got, gotMessage, testCase.want, testCase.wantMessage)
			}
		})
	}
}

func TestFromStatus(t *testing.T) {
	t.Parallel()

	type args struct {
		status int
		err    error
	}

	cases := map[string]struct {
		args    args
		wantErr error
	}{
		"no error": {
			args{},
			nil,
		},
		"bad": {
			args{
				status: http.StatusBadRequest,
				err:    errors.New("failure"),
			},
			errors.New("failure\ninvalid"),
		},
		"unauthorized": {
			args{
				status: http.StatusUnauthorized,
				err:    errors.New("failure"),
			},
			errors.New("failure\nunauthorized"),
		},
		"forbidden": {
			args{
				status: http.StatusForbidden,
				err:    errors.New("failure"),
			},
			errors.New("failure\nforbidden"),
		},
		"not found": {
			args{
				status: http.StatusNotFound,
				err:    errors.New("failure"),
			},
			errors.New("failure\nnot found"),
		},
		"not allowed": {
			args{
				status: http.StatusMethodNotAllowed,
				err:    errors.New("failure"),
			},
			errors.New("failure\nmethod not allowed"),
		},
		"internal error": {
			args{
				status: http.StatusInternalServerError,
				err:    errors.New("failure"),
			},
			errors.New("failure\ninternal error"),
		},
		"unknown": {
			args{
				status: http.StatusTeapot,
				err:    errors.New("failure"),
			},
			errors.New("failure"),
		},
	}

	for intention, testCase := range cases {
		t.Run(intention, func(t *testing.T) {
			t.Parallel()

			gotErr := FromStatus(testCase.args.status, testCase.args.err)

			failed := false

			switch {
			case
				testCase.wantErr == nil && gotErr != nil,
				testCase.wantErr != nil && gotErr == nil,
				testCase.wantErr != nil && !strings.Contains(gotErr.Error(), testCase.wantErr.Error()):
				failed = true
			}

			if failed {
				t.Errorf("FromStatus() = `%s`, want `%s`", gotErr, testCase.wantErr)
			}
		})
	}
}

func TestFromResponse(t *testing.T) {
	t.Parallel()

	type args struct {
		resp *http.Response
		err  error
	}

	cases := map[string]struct {
		args    args
		want    bool
		wantErr error
	}{
		"empty": {
			args{
				err: errors.New("failure"),
			},
			false,
			errors.New("failure"),
		},
		"valid": {
			args{
				resp: &http.Response{
					StatusCode: http.StatusMethodNotAllowed,
				},
				err: errors.New("failure"),
			},
			false,
			errors.New("failure\nmethod not allowed"),
		},
	}

	for intention, testCase := range cases {
		t.Run(intention, func(t *testing.T) {
			t.Parallel()

			gotErr := FromResponse(testCase.args.resp, testCase.args.err)

			failed := false

			switch {
			case
				testCase.wantErr == nil && gotErr != nil,
				testCase.wantErr != nil && gotErr == nil,
				testCase.wantErr != nil && !strings.Contains(gotErr.Error(), testCase.wantErr.Error()):
				failed = true
			}

			if failed {
				t.Errorf("FromResponse() = `%s`, want `%s`", gotErr, testCase.wantErr)
			}
		})
	}
}
