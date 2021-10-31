package request

import (
	"crypto/sha512"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/ViBiOh/httputils/v4/pkg/model"
)

func TestAddSignature(t *testing.T) {
	type args struct {
		keyID   string
		secret  []byte
		payload []byte
	}

	cases := []struct {
		intention         string
		args              args
		wantDigest        string
		wantAuthorization string
	}{
		{
			"simple",
			args{
				keyID:   "test",
				secret:  []byte(`password`),
				payload: []byte(`Hello World`),
			},
			"SHA-512=2c74fd17edafd80e8447b0d46741ee243b7eb74dd2149a0ab1b9246fb30382f27e853d8585719e0e67cbda0daa8f51671064615d645ae27acb15bfb1447f459b",
			`Signature keyId="test", algorithm="hs2019", headers="(request-target) digest", signature="5lf5ogggfJ1LXJciRS2BscNtMnYHWDOr2myJ9TJyZnu+37EXUpmchhl6LxyzU0bfpqAloLFEFw+1NEBSgNC+lQ=="`,
		},
	}

	for _, tc := range cases {
		t.Run(tc.intention, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/", nil)

			AddSignature(req, tc.args.keyID, tc.args.secret, tc.args.payload)

			if got := req.Header.Get("Digest"); got != tc.wantDigest {
				t.Errorf("AddSignature() = `%s`, want `%s`", got, tc.wantDigest)
			}
			if got := strings.Join(req.Header.Values("Authorization"), ", "); got != tc.wantAuthorization {
				t.Errorf("AddSignature() = `%s`, want `%s`", got, tc.wantAuthorization)
			}
		})
	}
}

func TestValidateSignature(t *testing.T) {
	type args struct {
		req    *http.Request
		secret []byte
	}

	reqNoHeader := httptest.NewRequest(http.MethodGet, "/", nil)

	reqWrongDigest := httptest.NewRequest(http.MethodGet, "/", strings.NewReader("Hello World"))
	reqWrongDigest.Header.Add("Digest", fmt.Sprintf("SHA-512=%x", sha512.Sum512([]byte(`Hello World!`))))

	reqNoAuth := httptest.NewRequest(http.MethodGet, "/", strings.NewReader("Hello World"))
	reqNoAuth.Header.Add("Digest", fmt.Sprintf("SHA-512=%x", sha512.Sum512([]byte(`Hello World`))))

	reqNoHeaders := httptest.NewRequest(http.MethodGet, "/", strings.NewReader("Hello World"))
	reqNoHeaders.Header.Add("Digest", fmt.Sprintf("SHA-512=%x", sha512.Sum512([]byte(`Hello World`))))
	reqNoHeaders.Header.Add("Authorization", "")

	reqNoSignature := httptest.NewRequest(http.MethodGet, "/", strings.NewReader("Hello World"))
	reqNoSignature.Header.Add("Digest", fmt.Sprintf("SHA-512=%x", sha512.Sum512([]byte(`Hello World`))))
	reqNoSignature.Header.Add("Authorization", `headers="(request-target) digest"`)

	reqInvalidSignature := httptest.NewRequest(http.MethodGet, "/", strings.NewReader("Hello World"))
	reqInvalidSignature.Header.Add("Digest", fmt.Sprintf("SHA-512=%x", sha512.Sum512([]byte(`Hello World`))))
	reqInvalidSignature.Header.Add("Authorization", `headers="(request-target) digest"`)
	reqInvalidSignature.Header.Add("Authorization", `signature="$"`)

	reqInvalidSecret := httptest.NewRequest(http.MethodGet, "/", strings.NewReader("Hello World"))
	reqInvalidSecret.Header.Add("Digest", fmt.Sprintf("SHA-512=%x", sha512.Sum512([]byte(`Hello World`))))
	reqInvalidSecret.Header.Add("Authorization", `headers="(request-target) digest"`)
	reqInvalidSecret.Header.Add("Authorization", `signature="5lf5ogggfJ1LXJciRS2BscNtMnYHWDOr2myJ9TJyZnu+37EXUpmchhl6LxyzU0bfpqAloLFEFw+1NEBSgNC+lQ=="`)

	cases := []struct {
		intention string
		args      args
		want      bool
		wantErr   error
	}{
		{
			"no header",
			args{
				req:    reqNoHeader,
				secret: []byte(`password`),
			},
			false,
			model.ErrInvalid,
		},
		{
			"invalid digest",
			args{
				req:    reqWrongDigest,
				secret: []byte(`password`),
			},
			false,
			model.ErrInvalid,
		},
		{
			"no authorization",
			args{
				req:    reqNoAuth,
				secret: []byte(`password`),
			},
			false,
			model.ErrInvalid,
		},
		{
			"no headers",
			args{
				req:    reqNoHeaders,
				secret: []byte(`password`),
			},
			false,
			model.ErrInvalid,
		},
		{
			"no signature",
			args{
				req:    reqNoSignature,
				secret: []byte(`password`),
			},
			false,
			model.ErrInvalid,
		},
		{
			"invalid signature",
			args{
				req:    reqInvalidSignature,
				secret: []byte(`password`),
			},
			false,
			model.ErrInvalid,
		},
		{
			"invalid secret",
			args{
				req:    reqInvalidSecret,
				secret: []byte(`passwor`),
			},
			false,
			nil,
		},
		{
			"invalid secret",
			args{
				req:    reqInvalidSecret,
				secret: []byte(`password`),
			},
			true,
			nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.intention, func(t *testing.T) {
			got, gotErr := ValidateSignature(tc.args.req, tc.args.secret)

			failed := false

			if tc.wantErr == nil && gotErr != nil {
				failed = true
			} else if tc.wantErr != nil && gotErr == nil {
				failed = true
			} else if tc.wantErr != nil && !strings.Contains(gotErr.Error(), tc.wantErr.Error()) {
				failed = true
			} else if got != tc.want {
				failed = true
			}

			if failed {
				t.Errorf("ValidateSignature() = (%t, `%s`), want (%t, `%s`)", got, gotErr, tc.want, tc.wantErr)
			}
		})
	}
}
