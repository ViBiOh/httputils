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
	URL       string
	UserAgent string
}

func Flags(fs *flag.FlagSet, prefix string, overrides ...flags.Override) *Config {
	var config Config

	flags.New("Url", "URL to check").Prefix(prefix).DocPrefix("alcotest").StringVar(fs, &config.URL, "", overrides)
	flags.New("UserAgent", "User-Agent for check").Prefix(prefix).DocPrefix("alcotest").StringVar(fs, &config.UserAgent, defaultUserAgent, overrides)

	return &config
}

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
		return status, err
	}

	if status <= http.StatusBadRequest {
		return status, err
	}

	return status, fmt.Errorf("HTTP/%d", status)
}

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

func DoAndExit(config *Config) {
	if len(config.URL) == 0 {
		return
	}

	if err := Do(config.URL, config.UserAgent); err != nil {
		fmt.Println(err)
		exitFunc(1)

		return
	}

	exitFunc(0)
}
