package alcotest

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/ViBiOh/httputils/v3/pkg/flags"
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
		url:       flags.New(prefix, "alcotest").Name("Url").Default("").Label("URL to check").ToString(fs),
		userAgent: flags.New(prefix, "alcotest").Name("UserAgent").Default("Golang alcotest").Label("User-Agent for check").ToString(fs),
	}
}

// GetStatusCode return status code of a GET on given url
func GetStatusCode(url, userAgent string) (int, error) {
	r, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return 0, err
	}

	if userAgent != "" {
		r.Header.Set("User-Agent", userAgent)
	}

	resp, err := httpClient.Do(r)
	if err != nil {
		return 0, err
	}

	if closeErr := resp.Body.Close(); closeErr != nil {
		if err == nil {
			return 0, closeErr
		}

		return 0, fmt.Errorf("%s: %w", err, closeErr)
	}

	return resp.StatusCode, nil
}

// Do test status code of given URL
func Do(url, userAgent string) error {
	statusCode, err := GetStatusCode(url, userAgent)
	if err != nil {
		return err
	}

	if statusCode > http.StatusNoContent {
		return fmt.Errorf("alcotest failed: HTTP/%d", statusCode)
	}

	return nil
}

// DoAndExit test status code of given URL (if present) and exit program with correct status
func DoAndExit(config Config) {
	url := strings.TrimSpace(*config.url)
	if url == "" {
		return
	}

	if err := Do(url, strings.TrimSpace(*config.userAgent)); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	os.Exit(0)
}
