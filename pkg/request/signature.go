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
	"strconv"
	"strings"
	"time"

	"github.com/ViBiOh/httputils/v4/pkg/model"
)

const (
	authorizationHeader = "Authorization"
	requestTargetHeader = "(request-target)"
	createdHeader       = "(created)"
	headerSeparator     = ": "
)

// AddSignature add Authorization header based on content signature based on https://datatracker.ietf.org/doc/html/draft-cavage-http-signatures-12
func AddSignature(r *http.Request, created time.Time, keyID string, secret, payload []byte) {
	digest := fmt.Sprintf("SHA-512=%x", sha512.Sum512(payload))
	r.Header.Add("Digest", digest)

	createdValue := created.Unix()

	r.Header.Add("Date", created.Format(time.RFC3339))

	r.Header.Add(authorizationHeader, fmt.Sprintf(`Signature keyId="%s"`, keyID))
	r.Header.Add(authorizationHeader, `algorithm="hs2019"`)
	r.Header.Add(authorizationHeader, `created=`+strconv.FormatInt(createdValue, 10))
	r.Header.Add(authorizationHeader, `headers="(request-target) (created) digest"`)
	signature := signContent(secret, buildSignatureString(r, []string{requestTargetHeader, createdHeader, "digest"}, createdValue))
	r.Header.Add(authorizationHeader, fmt.Sprintf(`signature="%s"`, base64.StdEncoding.EncodeToString(signature)))
}

// ValidateSignature check Authorization header based on content based on https://datatracker.ietf.org/doc/html/draft-cavage-http-signatures-12
func ValidateSignature(r *http.Request, secret []byte) (bool, error) {
	body, err := ReadBodyRequest(r)
	if err != nil {
		return false, fmt.Errorf("read body: %w", err)
	}

	r.Body = io.NopCloser(bytes.NewBuffer(body))

	if fmt.Sprintf("SHA-512=%x", sha512.Sum512(body)) != r.Header.Get("Digest") {
		return false, model.WrapInvalid(errors.New("SHA-512 signature of body doesn't match"))
	}

	signatureString, signature, err := parseAuthorizationHeader(r)
	if err != nil {
		return false, model.WrapInvalid(fmt.Errorf("parse authorization header: %w", err))
	}

	return hmac.Equal(signContent(secret, signatureString), signature), nil
}

func signContent(secret, content []byte) []byte {
	hash := hmac.New(sha512.New, secret)
	hash.Write(content)

	return hash.Sum(nil)
}

func parseAuthorizationHeader(r *http.Request) ([]byte, []byte, error) {
	var headers, algorithm, rawSignature string
	var created int64
	var err error

	for _, value := range r.Header.Values(authorizationHeader) {
		if strings.HasPrefix(value, "headers=") {
			headers = strings.TrimPrefix(value, "headers=")
		} else if strings.HasPrefix(value, "created=") {
			rawCreated := strings.TrimPrefix(value, "created=")

			created, err = strconv.ParseInt(rawCreated, 10, 64)
			if err != nil {
				return nil, nil, fmt.Errorf(createdHeader+" is not an integer: %w", err)
			}
		} else if strings.HasPrefix(value, "signature=") {
			rawSignature = strings.TrimPrefix(value, "signature=")
		} else if strings.HasPrefix(value, "algorithm=") {
			algorithm = strings.Trim(strings.TrimPrefix(value, "algorithm="), `"`)
		}
	}

	if len(headers) == 0 {
		headers = createdHeader
	}

	if len(rawSignature) == 0 {
		return nil, nil, errors.New("no signature section found in Authorization")
	}

	if strings.Contains(headers, createdHeader) && (strings.HasPrefix(algorithm, "rsa") || strings.HasPrefix(algorithm, "hmac") || strings.HasPrefix(algorithm, "ecdsa")) {
		return nil, nil, errors.New("`created` header is incompatible with algorithm")
	}

	signature, err := base64.StdEncoding.DecodeString(strings.Trim(rawSignature, `"`))
	if err != nil {
		return nil, nil, fmt.Errorf("decode base64 signature: %w", err)
	}

	return buildSignatureString(r, strings.Split(strings.Trim(headers, `"`), " "), created), signature, nil
}

func buildSignatureString(r *http.Request, parts []string, created int64) []byte {
	var signatureString bytes.Buffer

	for i, header := range parts {
		if i != 0 {
			signatureString.WriteString("\n")
		}

		switch header {
		case requestTargetHeader:
			signatureString.WriteString(requestTargetHeader)
			signatureString.WriteString(headerSeparator)
			signatureString.WriteString(strings.ToLower(fmt.Sprintf("%s %s", r.Method, r.URL.Path)))
		case createdHeader:
			signatureString.WriteString(createdHeader)
			signatureString.WriteString(headerSeparator)
			signatureString.WriteString(strconv.FormatInt(created, 10))
		default:
			signatureString.WriteString(strings.ToLower(header))
			signatureString.WriteString(headerSeparator)

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
