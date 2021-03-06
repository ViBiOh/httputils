package alcotest

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/ViBiOh/httputils/v4/pkg/flags"
	"github.com/ViBiOh/httputils/v4/pkg/request"
)

var (
	httpClient = http.Client{
		Timeout: 5 * time.Second,
	}

	exitFunc = os.Exit
)

// Config of package
type Config struct {
	url       *string
	userAgent *string
}

// Flags adds flags for configuring package
func Flags(fs *flag.FlagSet, prefix string, overrides ...flags.Override) Config {
	return Config{
		url:       flags.New(prefix, "alcotest").Name("Url").Default(flags.Default("Url", "", overrides)).Label("URL to check").ToString(fs),
		userAgent: flags.New(prefix, "alcotest").Name("UserAgent").Default(flags.Default("UserAgent", "Alcotest", overrides)).Label("User-Agent for check").ToString(fs),
	}
}

// GetStatusCode return status code of a GET on given url
func GetStatusCode(url, userAgent string) (int, error) {
	resp, err := request.New().Get(url).Header("User-Agent", userAgent).WithClient(httpClient).Send(context.Background(), nil)
	if resp == nil {
		return 0, err
	}

	return resp.StatusCode, err
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
	if len(url) == 0 {
		return
	}

	if err := Do(url, strings.TrimSpace(*config.userAgent)); err != nil {
		fmt.Println(err)
		exitFunc(1)
		return
	}

	exitFunc(0)
}
