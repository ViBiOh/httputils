package httputils

import (
	"bytes"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"time"
)

const clientTimeout = 30 * time.Second

var httpClient = http.Client{Timeout: clientTimeout}

var httpClientSkipTLS = http.Client{
	Timeout: clientTimeout,
	Transport: &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	},
}

func doAndRead(request *http.Request, skipTLSVerify bool) ([]byte, error) {
	client := httpClient
	if skipTLSVerify {
		client = httpClientSkipTLS
	}

	response, err := client.Do(request)
	if err != nil {
		return nil, fmt.Errorf(`Error while sending data: %v`, err)
	}
	responseBody, err := ReadBody(response.Body)

	if response.StatusCode >= http.StatusBadRequest {
		return nil, fmt.Errorf(`Error status %d: %s`, response.StatusCode, responseBody)
	}

	if err != nil {
		return nil, fmt.Errorf(`Error while reading body: %v`, err)
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
func GetBody(url string, headers map[string]string, skipTLSVerify bool) ([]byte, error) {
	request, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf(`Error while creating request: %v`, err)
	}

	for key, value := range headers {
		request.Header.Add(key, value)
	}

	return doAndRead(request, skipTLSVerify)
}

// PostJSONBody post given interface to URL with optional credential supplied
func PostJSONBody(url string, body interface{}, headers map[string]string, skipTLSVerify bool) ([]byte, error) {
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf(`Error while marshalling body: %v`, err)
	}

	request, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf(`Error while creating request: %v`, err)
	}

	for key, value := range headers {
		request.Header.Add(key, value)
	}

	request.Header.Add(`Content-Type`, `application/json`)

	return doAndRead(request, skipTLSVerify)
}
