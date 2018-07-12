package alcotest

import (
	"crypto/tls"
	"flag"
	"fmt"
	"net/http"
	"os"
	"strings"
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

// Flags add flags for given prefix
func Flags(prefix string) map[string]*string {
	return map[string]*string{
		`url`: flag.String(tools.ToCamel(fmt.Sprintf(`%sUrl`, prefix)), ``, `[health] URL to check`),
		`userAgent`: flag.String(tools.ToCamel(fmt.Sprintf(`%sUserAgent`, prefix)), `Golang alcotest`, `[health] User-Agent used`),
	}
}

// GetStatusCode return status code of a GET on given url
func GetStatusCode(url, userAgent string) (status int, err error) {
	r, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		err = fmt.Errorf(`Error while creating request: %v`, err)
		return
	}

	if userAgent != `` {
		r.Header.Set(`User-Agent`, userAgent)
	}

	var response *http.Response

	response, err = httpClient.Do(r)
	if response != nil {
		status = response.StatusCode

		defer func() {
			if closeErr := response.Body.Close(); closeErr != nil {
				err = fmt.Errorf(`%s, and also error while closing response: %v`, err, closeErr)
			}
		}()
	}

	return
}

// Do test status code of given URL
func Do(url, userAgent string) error {
	statusCode, err := GetStatusCode(url, userAgent)
	if err != nil {
		return fmt.Errorf(`Unable to blow in balloon: %v`, err)
	}

	if statusCode != http.StatusOK {
		return fmt.Errorf(`Alcotest failed: HTTP/%d`, statusCode)
	}

	return nil
}

// DoAndExit test status code of given URL (if present) and exit program with correct status
func DoAndExit(config map[string]*string) {
	url := strings.TrimSpace(*config[`url`])
	userAgent := strings.TrimSpace(*config[`userAgent`])

	if url != `` {
		if err := Do(url, userAgent); err != nil {
			fmt.Print(err)
			os.Exit(1)
		} else {
			os.Exit(0)
		}
	}
}
