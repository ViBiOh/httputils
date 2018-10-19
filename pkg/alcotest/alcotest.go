package alcotest

import (
	"crypto/tls"
	"flag"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/ViBiOh/httputils/pkg/errors"
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

// Flags add flags for given prefix
func Flags(prefix string) map[string]*string {
	return map[string]*string{
		`url`:       flag.String(tools.ToCamel(fmt.Sprintf(`%sUrl`, prefix)), ``, `[health] URL to check`),
		`userAgent`: flag.String(tools.ToCamel(fmt.Sprintf(`%sUserAgent`, prefix)), `Golang alcotest`, `[health] User-Agent for check`),
	}
}

// GetStatusCode return status code of a GET on given url
func GetStatusCode(url, userAgent string) (int, error) {
	r, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return 0, errors.WithStack(err)
	}

	if userAgent != `` {
		r.Header.Set(`User-Agent`, userAgent)
	}

	response, err := httpClient.Do(r)
	if response != nil {
		if closeErr := response.Body.Close(); closeErr != nil {
			if err == nil {
				return 0, errors.WithStack(closeErr)
			}

			return 0, errors.New(`%v, and also %v`, err, closeErr)
		}

		return response.StatusCode, nil
	}

	if err != nil {
		return 0, errors.WithStack(err)
	}

	return 0, nil
}

// Do test status code of given URL
func Do(url, userAgent string) error {
	statusCode, err := GetStatusCode(url, userAgent)
	if err != nil {
		return err
	}

	if statusCode != http.StatusOK {
		return errors.New(`alcotest failed: HTTP/%d`, statusCode)
	}

	return nil
}

// DoAndExit test status code of given URL (if present) and exit program with correct status
func DoAndExit(config map[string]*string) {
	url := strings.TrimSpace(*config[`url`])
	userAgent := strings.TrimSpace(*config[`userAgent`])

	if url != `` {
		if err := Do(url, userAgent); err != nil {
			fmt.Printf(`%+v`, err)
			os.Exit(1)
		} else {
			os.Exit(0)
		}
	}
}
