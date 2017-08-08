package httputils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"time"
)

var httpClient = http.Client{Timeout: 30 * time.Second}

func doAndRead(request *http.Request) ([]byte, error) {
	response, err := httpClient.Do(request)
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

func addAuthorization(request *http.Request, authorization string) {
	if authorization != `` {
		request.Header.Add(`Authorization`, authorization)
	}
}

// ReadBody return content of a body request (defined as a ReadCloser)
func ReadBody(body io.ReadCloser) ([]byte, error) {
	defer body.Close()
	return ioutil.ReadAll(body)
}

// GetBody return body of given URL or error if something goes wrong
func GetBody(url string, authorization string) ([]byte, error) {
	request, err := http.NewRequest(`GET`, url, nil)
	if err != nil {
		return nil, fmt.Errorf(`Error while creating request: %v`, err)
	}

	addAuthorization(request, authorization)

	return doAndRead(request)
}

// PostJSONBody post given interface to URL with optional credential supplied
func PostJSONBody(url string, body interface{}, authorization string) ([]byte, error) {
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf(`Error while marshalling body: %v`, err)
	}

	request, err := http.NewRequest(`POST`, url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf(`Error while creating request: %v`, err)
	}

	addAuthorization(request, authorization)
	request.Header.Add(`Content-Type`, `application/json`)

	return doAndRead(request)
}
