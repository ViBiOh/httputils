package alcotest

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/ViBiOh/flags"
	"github.com/ViBiOh/httputils/v4/pkg/request"
)

var (
	httpClient       = request.CreateClient(5*time.Second, request.NoRedirection)
	defaultURL       = "http://localhost:1080/health"
	defaultUserAgent = "Alcotest"
	defaultHeader    = http.Header{}
	exitFunc         = os.Exit

	req *http.Request
)

func init() {
	req, _ = http.NewRequest(http.MethodGet, defaultURL, nil)
	defaultHeader.Set("User-Agent", defaultUserAgent)
}

type Config struct {
	url       *string
	userAgent *string
}

func Flags(fs *flag.FlagSet, prefix string, overrides ...flags.Override) Config {
	return Config{
		url:       flags.String(fs, prefix, "alcotest", "Url", "URL to check", "", overrides),
		userAgent: flags.String(fs, prefix, "alcotest", "UserAgent", "User-Agent for check", defaultUserAgent, overrides),
	}
}

// GetStatusCode return status code of a GET on given url.
func GetStatusCode(url, userAgent string) (status int, err error) {
	statusReq := req

	if url != defaultURL {
		statusReq, err = http.NewRequest(http.MethodGet, url, nil)
		if err != nil {
			return 0, fmt.Errorf("create request: %w", err)
		}
		statusReq.Header = defaultHeader
	}

	if userAgent != defaultUserAgent {
		statusReq.Header.Set("User-Agent", userAgent)
	}

	var resp *http.Response
	resp, err = httpClient.Do(statusReq)
	if err != nil {
		return 0, fmt.Errorf("perform request: %w", err)
	}

	status = resp.StatusCode

	if err = request.DiscardBody(resp.Body); err != nil {
		return
	}

	if status <= http.StatusBadRequest {
		return
	}

	err = fmt.Errorf("HTTP/%d", status)

	return
}

// Do test status code of given URL.
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

// DoAndExit test status code of given URL (if present) and exit program with correct status.
func DoAndExit(config Config) {
	url := *config.url
	if len(url) == 0 {
		return
	}

	if err := Do(url, *config.userAgent); err != nil {
		fmt.Println(err)
		exitFunc(1)

		return
	}

	exitFunc(0)
}
