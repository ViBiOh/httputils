package httputils

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"time"
)

// ForwardedForHeader that proxy uses to fill
const ForwardedForHeader = `X-Forwarded-For`

var httpClient = http.Client{Timeout: 30 * time.Second}

func doAndRead(request *http.Request) ([]byte, error) {
	response, err := httpClient.Do(request)
	if err != nil {
		if response != nil {
			response.Body.Close()
		}
		return nil, fmt.Errorf(`Error while processing request: %v`, err)
	}

	responseBody, err := ReadBody(response.Body)
	if err != nil {
		return nil, fmt.Errorf(`Error while reading body: %v`, err)
	}

	if response.StatusCode >= http.StatusBadRequest {
		return responseBody, fmt.Errorf(`Error status %d`, response.StatusCode)
	}

	return responseBody, nil
}

// GetBasicAuth generates Basic Auth for given username and password
func GetBasicAuth(username string, password string) string {
	return `Basic ` + base64.StdEncoding.EncodeToString([]byte(username+`:`+password))
}

// ReadBody return content of a body request (defined as a ReadCloser)
func ReadBody(body io.ReadCloser) ([]byte, error) {
	defer body.Close()
	return ioutil.ReadAll(body)
}

// GetBody return body of given URL or error if something goes wrong
func GetBody(url string, headers map[string]string) ([]byte, error) {
	request, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf(`Error while creating request: %v`, err)
	}

	for key, value := range headers {
		request.Header.Add(key, value)
	}

	return doAndRead(request)
}

// MethodBody send given method with given content to URL with optional headers supplied
func MethodBody(url string, body []byte, headers map[string]string, method string) ([]byte, error) {
	request, err := http.NewRequest(method, url, bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf(`Error while creating request: %v`, err)
	}

	for key, value := range headers {
		request.Header.Add(key, value)
	}

	return doAndRead(request)
}

// PostBody post given content to URL with optional headers supplied
func PostBody(url string, body []byte, headers map[string]string) ([]byte, error) {
	return MethodBody(url, body, headers, http.MethodPost)
}

// PostJSONBody post given interface{} to URL with optional headers supplied
func PostJSONBody(url string, body interface{}, headers map[string]string) ([]byte, error) {
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf(`Error while marshalling body: %v`, err)
	}

	if headers == nil {
		headers = make(map[string]string)
	}
	headers[`Content-Type`] = `application/json`

	return PostBody(url, jsonBody, headers)
}

// GetIP give remote IP
func GetIP(r *http.Request) (ip string) {
	ip = r.Header.Get(ForwardedForHeader)
	if ip == `` {
		ip = r.RemoteAddr
	}

	return
}
