package alcotest

import (
	"crypto/tls"
	"flag"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/ViBiOh/httputils/pkg/tools"
)

var httpClient = http.Client{
	Timeout: 5 * time.Second,
	Transport: &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	},
}

// GetStatusCode return status code of a GET on given url
func GetStatusCode(url string) (int, error) {
	request, err := http.NewRequest(`GET`, url, nil)
	if err != nil {
		return 0, err
	}
	request.Header.Set(`X-Forwarded-For`, `alcotest`)

	response, err := httpClient.Do(request)
	if response != nil {
		if closeErr := response.Body.Close(); closeErr != nil {
			err = fmt.Errorf(`%s, and also error while closing response: %v`, err, closeErr)
		}
	}
	if err != nil {
		return 0, err
	}

	return response.StatusCode, nil
}

// Flags add flags for given prefix
func Flags(prefix string) map[string]*string {
	return map[string]*string{
		`url`: flag.String(tools.ToCamel(fmt.Sprintf(`%s%s`, prefix, `Url`)), ``, `[health] URL to check`),
	}
}

// Do test status code of given URL
func Do(url string) error {
	if statusCode, err := GetStatusCode(url); err != nil {
		return fmt.Errorf(`Unable to blow in ballon: %v`, err)
	} else if statusCode != http.StatusOK {
		return fmt.Errorf(`Alcotest failed: HTTP/%d`, statusCode)
	}

	return nil
}

// DoAndExit test status code of given URL (if present) and exit program with correct status
func DoAndExit(config map[string]*string) {
	url := *config[`url`]

	if url != `` {
		if err := Do(url); err != nil {
			fmt.Print(err)
			os.Exit(1)
		} else {
			os.Exit(0)
		}
	}
}
