package request

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha512"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/ViBiOh/httputils/v4/pkg/model"
)

const (
	authorizationHeader = "Authorization"
)

// AddSignature add Authorization header based on content signature based on https://datatracker.ietf.org/doc/html/draft-cavage-http-signatures-12
func AddSignature(r *http.Request, keyID string, secret, payload []byte) {
	digest := fmt.Sprintf("SHA-512=%x", sha512.Sum512(payload))
	r.Header.Add("Digest", digest)

	r.Header.Add(authorizationHeader, fmt.Sprintf(`Signature keyId="%s"`, keyID))
	r.Header.Add(authorizationHeader, `algorithm="hs2019"`)
	r.Header.Add(authorizationHeader, `headers="(request-target) digest"`)
	signature := signContent(secret, buildSignatureString(r, []string{"(request-target)", "digest"}))
	r.Header.Add(authorizationHeader, fmt.Sprintf(`signature="%s"`, base64.StdEncoding.EncodeToString(signature)))
}

// ValidateSignature check Authorization header based on content based on https://datatracker.ietf.org/doc/html/draft-cavage-http-signatures-12
func ValidateSignature(r *http.Request, secret []byte) (bool, error) {
	body, err := ReadBodyRequest(r)
	if err != nil {
		return false, fmt.Errorf("read body: %s", err)
	}

	r.Body = io.NopCloser(bytes.NewBuffer(body))

	if fmt.Sprintf("SHA-512=%x", sha512.Sum512(body)) != r.Header.Get("Digest") {
		return false, model.WrapInvalid(errors.New("SHA-512 signature of body doesn't match"))
	}

	signatureString, signature, err := parseAuthorizationHeader(r)
	if err != nil {
		return false, model.WrapInvalid(fmt.Errorf("parse authorization header: %s", err))
	}

	return hmac.Equal(signContent(secret, signatureString), signature), nil
}

func signContent(secret, content []byte) []byte {
	hash := hmac.New(sha512.New, secret)
	hash.Write(content)
	return hash.Sum(nil)
}

func parseAuthorizationHeader(r *http.Request) ([]byte, []byte, error) {
	var rawHeaders, rawSignature string

	for _, value := range r.Header.Values(authorizationHeader) {
		if strings.HasPrefix(value, "headers=") {
			rawHeaders = value
		} else if strings.HasPrefix(value, "signature=") {
			rawSignature = value
		}
	}

	if len(rawHeaders) == 0 {
		return nil, nil, errors.New("no headers section found in Authorization")
	}

	if len(rawSignature) == 0 {
		return nil, nil, errors.New("no signature section found in Authorization")
	}

	signature, err := base64.StdEncoding.DecodeString(strings.Trim(strings.TrimPrefix(rawSignature, "signature="), `"`))
	if err != nil {
		return nil, nil, fmt.Errorf("decode base64 signature: %s", err)
	}

	signatureString := buildSignatureString(r, strings.Split(strings.Trim(strings.TrimPrefix(rawHeaders, "headers="), `"`), " "))
	return signatureString, signature, nil
}

func buildSignatureString(r *http.Request, parts []string) []byte {
	var signatureString bytes.Buffer

	for i, header := range parts {
		if i != 0 {
			signatureString.WriteString("\n")
		}

		if header == "(request-target)" {
			signatureString.WriteString(strings.ToLower(fmt.Sprintf("%s %s", r.Method, r.URL.Path)))
		} else {
			signatureString.WriteString(strings.ToLower(header))
			signatureString.WriteString(": ")

			for j, value := range r.Header.Values(header) {
				if j != 0 {
					signatureString.WriteString(", ")
				}

				signatureString.WriteString(strings.TrimSpace(value))
			}
		}
	}

	return signatureString.Bytes()
}
