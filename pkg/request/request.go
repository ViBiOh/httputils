package request

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"

	"github.com/opentracing-contrib/go-stdlib/nethttp"
	opentracing "github.com/opentracing/opentracing-go"
)

const (
	// ForwardedForHeader that proxy uses to fill
	ForwardedForHeader = `X-Forwarded-For`

	// ContentTypeHeader value
	ContentTypeHeader = `Content-Type`
)

var defaultHTTPClient = http.Client{
	Timeout:   30 * time.Second,
	Transport: &nethttp.Transport{},
}

// DoAndReadWithClient execute read and return given request on given client
func DoAndReadWithClient(ctx context.Context, client http.Client, request *http.Request) ([]byte, error) {
	if ctx != nil {
		var netTracer *nethttp.Tracer

		request = request.WithContext(ctx)
		request, netTracer = nethttp.TraceRequest(opentracing.GlobalTracer(), request)
		defer netTracer.Finish()
	}

	response, err := client.Do(request)
	if err != nil {
		if response != nil {
			if closeErr := response.Body.Close(); closeErr != nil {
				err = fmt.Errorf(`, and also error while closing body: %v`, closeErr)
			}
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

func doAndRead(ctx context.Context, request *http.Request) ([]byte, error) {
	return DoAndReadWithClient(ctx, defaultHTTPClient, request)
}

// GetBasicAuth generates Basic Auth for given username and password
func GetBasicAuth(username string, password string) string {
	return fmt.Sprintf(`Basic %s`, base64.StdEncoding.EncodeToString([]byte(username+`:`+password)))
}

// ReadBody return content of a body request (defined as a ReadCloser)
func ReadBody(body io.ReadCloser) (_ []byte, err error) {
	defer func() {
		err = body.Close()
	}()
	return ioutil.ReadAll(body)
}

// Do send given method with given content to URL with optional headers supplied
func Do(ctx context.Context, url string, body []byte, headers http.Header, method string) ([]byte, error) {
	request, err := http.NewRequest(method, url, bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf(`Error while creating request: %v`, err)
	}
	request.Header = headers

	return doAndRead(ctx, request)
}

// DoJSON send given method with given interface{} as JSON to URL with optional headers supplied
func DoJSON(ctx context.Context, url string, body interface{}, headers http.Header, method string) ([]byte, error) {
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf(`Error while marshalling body: %v`, err)
	}

	if headers == nil {
		headers = http.Header{}
	}
	headers.Set(ContentTypeHeader, `application/json`)

	return Do(ctx, url, jsonBody, headers, method)
}

// Get send GET request to URL with optional headers supplied
func Get(ctx context.Context, url string, headers http.Header) ([]byte, error) {
	return Do(ctx, url, nil, headers, http.MethodGet)
}

// PostForm send form via POST with urlencoded data
func PostForm(ctx context.Context, url string, headers http.Header, data url.Values) ([]byte, error) {
	if headers == nil {
		headers = http.Header{}
	}
	headers.Set(ContentTypeHeader, `application/x-www-form-urlencoded`)

	return Do(ctx, url, []byte(data.Encode()), headers, http.MethodPost)
}

// SetIP set remote IP
func SetIP(r *http.Request, ip string) {
	r.Header.Set(ForwardedForHeader, ip)
}

// GetIP give remote IP
func GetIP(r *http.Request) (ip string) {
	ip = r.Header.Get(ForwardedForHeader)
	if ip == `` {
		ip = r.RemoteAddr
	}

	return
}
