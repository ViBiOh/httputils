package alcotest

import (
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
}

// Config of package
type Config struct {
	url       *string
	userAgent *string
}

// Flags adds flags for configuring package
func Flags(fs *flag.FlagSet, prefix string) Config {
	return Config{
		url:       tools.NewFlag(prefix, "alcotest").Name("Url").Default("").Label("URL to check").ToString(fs),
		userAgent: tools.NewFlag(prefix, "alcotest").Name("UserAgent").Default("Golang alcotest").Label("User-Agent for check").ToString(fs),
	}
}

// GetStatusCode return status code of a GET on given url
func GetStatusCode(url, userAgent string) (int, error) {
	r, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return 0, errors.WithStack(err)
	}

	if userAgent != "" {
		r.Header.Set("User-Agent", userAgent)
	}

	response, err := httpClient.Do(r)
	if err != nil {
		return 0, errors.WithStack(err)
	}

	if closeErr := response.Body.Close(); closeErr != nil {
		if err == nil {
			return 0, errors.WithStack(closeErr)
		}

		return 0, errors.Wrap(closeErr, "%#v", err)
	}

	return response.StatusCode, nil
}

// Do test status code of given URL
func Do(url, userAgent string) error {
	statusCode, err := GetStatusCode(url, userAgent)
	if err != nil {
		return err
	}

	if statusCode > http.StatusNoContent {
		return errors.New("alcotest failed: HTTP/%d", statusCode)
	}

	return nil
}

// DoAndExit test status code of given URL (if present) and exit program with correct status
func DoAndExit(config Config) {
	url := strings.TrimSpace(*config.url)
	userAgent := strings.TrimSpace(*config.userAgent)

	if url != "" {
		if err := Do(url, userAgent); err != nil {
			fmt.Printf("%#v", err)
			os.Exit(1)
		} else {
			os.Exit(0)
		}
	}
}
