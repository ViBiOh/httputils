package request

import (
	"crypto/sha512"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"slices"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/ViBiOh/httputils/v4/pkg/model"
	"github.com/stretchr/testify/assert"
)

func TestAddSignature(t *testing.T) {
	t.Parallel()

	type args struct {
		keyID   string
		secret  []byte
		payload []byte
	}

	cases := map[string]struct {
		args              args
		wantDigest        string
		wantAuthorization string
	}{
		"simple": {
			args{
				keyID:   "test",
				secret:  []byte(`password`),
				payload: []byte(`Hello World`),
			},
			"SHA-512=2c74fd17edafd80e8447b0d46741ee243b7eb74dd2149a0ab1b9246fb30382f27e853d8585719e0e67cbda0daa8f51671064615d645ae27acb15bfb1447f459b",
			`Signature keyId="test", algorithm="hs2019", created=1671913800, headers="(request-target) (created) digest", signature="yr9tUFlYgFvIHFVth7knuim3m7lrYcv3y1j7ltPlQu5TSaCwM2xlxNo1AVyVrTYEKVgk8t2Kcq3piScvhqqoZg=="`,
		},
	}

	for intention, testCase := range cases {
		t.Run(intention, func(t *testing.T) {
			t.Parallel()

			req := httptest.NewRequest(http.MethodGet, "/", nil)

			created := time.Date(2022, 12, 24, 20, 30, 0, 0, time.UTC)

			AddSignature(req, created, testCase.args.keyID, testCase.args.secret, testCase.args.payload)

			if got := req.Header.Get("Digest"); got != testCase.wantDigest {
				t.Errorf("AddSignature() = `%s`, want `%s`", got, testCase.wantDigest)
			}
			if got := strings.Join(req.Header.Values("Authorization"), ", "); got != testCase.wantAuthorization {
				t.Errorf("AddSignature() = `%s`, want `%s`", got, testCase.wantAuthorization)
			}
		})
	}
}

func signedRequest(path, body string, created time.Time, signedHeaders []string, secret []byte) *http.Request {
	req := httptest.NewRequest(http.MethodPost, path, strings.NewReader(body))
	req.Header.Add("Digest", fmt.Sprintf("SHA-512=%x", sha512.Sum512([]byte(body))))

	createdValue := created.Unix()
	signature := signContent(secret, buildSignatureString(req, signedHeaders, createdValue))

	req.Header.Add(authorizationHeader, `Signature keyId="test"`)
	req.Header.Add(authorizationHeader, `algorithm="hs2019"`)
	req.Header.Add(authorizationHeader, `created=`+strconv.FormatInt(createdValue, 10))

	if len(signedHeaders) != 0 {
		req.Header.Add(authorizationHeader, fmt.Sprintf(`headers="%s"`, strings.Join(signedHeaders, " ")))
	}

	req.Header.Add(authorizationHeader, fmt.Sprintf(`signature="%s"`, base64.StdEncoding.EncodeToString(signature)))

	return req
}

func TestValidateSignature(t *testing.T) {
	t.Parallel()

	secret := []byte(`SecureSymmetricKeyWithLongValueInIt`)
	now := time.Date(2022, 12, 24, 20, 30, 0, 0, time.UTC)
	maxAge := time.Minute
	allHeaders := []string{requestTargetHeader, createdHeader, "digest"}

	cases := map[string]struct {
		build   func() *http.Request
		want    bool
		wantErr error
	}{
		"valid": {
			func() *http.Request {
				return signedRequest("/signed", "Hello World", now, allHeaders, secret)
			},
			true,
			nil,
		},
		"valid at the oldest accepted instant": {
			func() *http.Request {
				return signedRequest("/signed", "Hello World", now.Add(-maxAge), allHeaders, secret)
			},
			true,
			nil,
		},
		"valid at the furthest accepted future": {
			func() *http.Request {
				return signedRequest("/signed", "Hello World", now.Add(maxAge), allHeaders, secret)
			},
			true,
			nil,
		},
		"valid without body": {
			func() *http.Request {
				return signedRequest("/signed", "", now, allHeaders, secret)
			},
			true,
			nil,
		},
		"replayed after the window": {
			func() *http.Request {
				return signedRequest("/signed", "Hello World", now.Add(-maxAge-time.Second), allHeaders, secret)
			},
			false,
			model.ErrInvalid,
		},
		"created too far in the future": {
			func() *http.Request {
				return signedRequest("/signed", "Hello World", now.Add(maxAge+time.Second), allHeaders, secret)
			},
			false,
			model.ErrInvalid,
		},
		"created missing": {
			func() *http.Request {
				req := signedRequest("/signed", "Hello World", now, allHeaders, secret)
				authorization := slices.DeleteFunc(req.Header.Values(authorizationHeader), func(value string) bool {
					return strings.HasPrefix(value, "created=")
				})
				req.Header.Del(authorizationHeader)

				for _, value := range authorization {
					req.Header.Add(authorizationHeader, value)
				}

				return req
			},
			false,
			model.ErrInvalid,
		},
		"digest not signed": {
			func() *http.Request {
				return signedRequest("/signed", "Hello World", now, []string{requestTargetHeader, createdHeader}, secret)
			},
			false,
			model.ErrInvalid,
		},
		"request target not signed": {
			func() *http.Request {
				return signedRequest("/signed", "Hello World", now, []string{createdHeader, "digest"}, secret)
			},
			false,
			model.ErrInvalid,
		},
		"created not signed": {
			func() *http.Request {
				return signedRequest("/signed", "Hello World", now, []string{requestTargetHeader, "digest"}, secret)
			},
			false,
			model.ErrInvalid,
		},
		"default signed headers": {
			func() *http.Request {
				return signedRequest("/signed", "Hello World", now, nil, secret)
			},
			false,
			model.ErrInvalid,
		},
		"request target spelled with another case does not count": {
			func() *http.Request {
				return signedRequest("/signed", "Hello World", now, []string{"(Request-Target)", createdHeader, "digest"}, secret)
			},
			false,
			model.ErrInvalid,
		},
		"signature reused on another path": {
			func() *http.Request {
				req := signedRequest("/signed", "Hello World", now, allHeaders, secret)
				req.URL.Path = "/other"

				return req
			},
			false,
			nil,
		},
		"signature reused on another method": {
			func() *http.Request {
				req := signedRequest("/signed", "Hello World", now, allHeaders, secret)
				req.Method = http.MethodDelete

				return req
			},
			false,
			nil,
		},
		"body replaced": {
			func() *http.Request {
				req := signedRequest("/signed", "Hello World", now, allHeaders, secret)
				req.Body = io.NopCloser(strings.NewReader("Hello World!"))

				return req
			},
			false,
			model.ErrInvalid,
		},
		"body too large": {
			func() *http.Request {
				return signedRequest("/signed", strings.Repeat("x", 2048), now, allHeaders, secret)
			},
			false,
			model.ErrInvalid,
		},
		"wrong secret": {
			func() *http.Request {
				return signedRequest("/signed", "Hello World", now, allHeaders, []byte(`passwor`))
			},
			false,
			nil,
		},
		"no authorization": {
			func() *http.Request {
				return httptest.NewRequest(http.MethodGet, "/", strings.NewReader("Hello World"))
			},
			false,
			model.ErrInvalid,
		},
		"empty authorization": {
			func() *http.Request {
				req := httptest.NewRequest(http.MethodGet, "/", strings.NewReader("Hello World"))
				req.Header.Add(authorizationHeader, "")

				return req
			},
			false,
			model.ErrInvalid,
		},
		"invalid signature encoding": {
			func() *http.Request {
				req := httptest.NewRequest(http.MethodGet, "/", strings.NewReader("Hello World"))
				req.Header.Add(authorizationHeader, fmt.Sprintf(`headers="%s"`, strings.Join(allHeaders, " ")))
				req.Header.Add(authorizationHeader, `signature="$"`)

				return req
			},
			false,
			model.ErrInvalid,
		},
	}

	for intention, testCase := range cases {
		t.Run(intention, func(t *testing.T) {
			t.Parallel()

			got, gotErr := validateSignature(testCase.build(), secret, 1024, maxAge, now)

			assert.Equal(t, testCase.want, got)

			if testCase.wantErr != nil {
				assert.ErrorIs(t, gotErr, testCase.wantErr)
			} else {
				assert.NoError(t, gotErr)
			}
		})
	}
}

func TestSignatureRoundTrip(t *testing.T) {
	t.Parallel()

	secret := []byte(`SecureSymmetricKeyWithLongValueInIt`)

	req := httptest.NewRequest(http.MethodPost, "/signed", strings.NewReader("Hello World"))
	AddSignature(req, time.Now(), "test", secret, []byte("Hello World"))

	got, err := ValidateSignature(req, secret, 1024, time.Minute)

	assert.NoError(t, err)
	assert.True(t, got)

	body, err := io.ReadAll(req.Body)

	assert.NoError(t, err)
	assert.Equal(t, "Hello World", string(body))
}
