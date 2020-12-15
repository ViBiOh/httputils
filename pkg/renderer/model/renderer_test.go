package model

import (
	"errors"
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
				err: WrapInvalid(errors.New("bad request")),
			},
			http.StatusBadRequest,
			"bad request",
		},
		{
			"not found",
			args{
				err: WrapNotFound(errors.New("unknown")),
			},
			http.StatusNotFound,
			"unknown",
		},
		{
			"method",
			args{
				err: WrapMethodNotAllowed(errors.New("not allowed")),
			},
			http.StatusMethodNotAllowed,
			"not allowed",
		},
		{
			"internal",
			args{
				err: WrapInternal(errors.New("boom")),
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
