package httputils

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func Test_BadRequest(t *testing.T) {
	var cases = []struct {
		err  error
		want string
	}{
		{
			fmt.Errorf(`BadRequest`),
			`BadRequest
`,
		},
	}

	for _, testCase := range cases {
		writer := httptest.NewRecorder()
		BadRequest(writer, testCase.err)

		if result := writer.Result().StatusCode; result != http.StatusBadRequest {
			t.Errorf(`badRequest(%v) = %v, want %v`, testCase.err, result, http.StatusBadRequest)
		}

		if result, _ := ReadBody(writer.Result().Body); string(result) != string(testCase.want) {
			t.Errorf(`badRequest(%v) = %v, want %v`, testCase.err, string(result), string(testCase.want))
		}
	}
}

func Test_Unauthorized(t *testing.T) {
	var cases = []struct {
		err  error
		want string
	}{
		{
			fmt.Errorf(`Unauthorized`),
			`Unauthorized
`,
		},
	}

	for _, testCase := range cases {
		writer := httptest.NewRecorder()
		Unauthorized(writer, testCase.err)

		if result := writer.Result().StatusCode; result != http.StatusUnauthorized {
			t.Errorf(`badRequest(%v) = %v, want %v`, testCase.err, result, http.StatusUnauthorized)
		}

		if result, _ := ReadBody(writer.Result().Body); string(result) != string(testCase.want) {
			t.Errorf(`unauthorized(%v) = %v, want %v`, testCase.err, string(result), string(testCase.want))
		}
	}
}

func Test_Forbidden(t *testing.T) {
	var cases = []struct {
	}{
		{},
	}

	for range cases {
		writer := httptest.NewRecorder()
		Forbidden(writer)

		if result := writer.Result().StatusCode; result != http.StatusForbidden {
			t.Errorf(`forbidden() = %v, want %v`, result, http.StatusForbidden)
		}
	}
}

func Test_InternalServerError(t *testing.T) {
	var cases = []struct {
		err  error
		want string
	}{
		{
			fmt.Errorf(`Internal server error`),
			`Internal server error
`,
		},
	}

	for _, testCase := range cases {
		writer := httptest.NewRecorder()
		InternalServerError(writer, testCase.err)

		if result := writer.Result().StatusCode; result != http.StatusInternalServerError {
			t.Errorf(`errorHandler(%v) = %v, want %v`, testCase.err, result, http.StatusInternalServerError)
		}
	}
}
