package cors

import (
	"flag"
	"net/http"
	"strconv"

	"github.com/ViBiOh/flags"
	"github.com/ViBiOh/httputils/v4/pkg/model"
)

var _ model.Middleware = App{}.Middleware

type App struct {
	origin      string
	headers     string
	methods     string
	exposes     string
	credentials string
}

type Config struct {
	origin      *string
	headers     *string
	methods     *string
	exposes     *string
	credentials *bool
}

func Flags(fs *flag.FlagSet, prefix string, overrides ...flags.Override) Config {
	return Config{
		origin:      flags.New("Origin", "Access-Control-Allow-Origin").Prefix(prefix).DocPrefix("cors").String(fs, "*", overrides),
		headers:     flags.New("Headers", "Access-Control-Allow-Headers").Prefix(prefix).DocPrefix("cors").String(fs, "Content-Type", overrides),
		methods:     flags.New("Methods", "Access-Control-Allow-Methods").Prefix(prefix).DocPrefix("cors").String(fs, http.MethodGet, overrides),
		exposes:     flags.New("Expose", "Access-Control-Expose-Headers").Prefix(prefix).DocPrefix("cors").String(fs, "", overrides),
		credentials: flags.New("Credentials", "Access-Control-Allow-Credentials").Prefix(prefix).DocPrefix("cors").Bool(fs, false, overrides),
	}
}

func New(config Config) App {
	return App{
		origin:      *config.origin,
		headers:     *config.headers,
		methods:     *config.methods,
		exposes:     *config.exposes,
		credentials: strconv.FormatBool(*config.credentials),
	}
}

func (a App) Middleware(next http.Handler) http.Handler {
	headers := http.Header{}

	if len(a.origin) != 0 {
		headers.Add("Access-Control-Allow-Origin", a.origin)
	}
	if len(a.headers) != 0 {
		headers.Add("Access-Control-Allow-Headers", a.headers)
	}
	if len(a.methods) != 0 {
		headers.Add("Access-Control-Allow-Methods", a.methods)
	}
	if len(a.exposes) != 0 {
		headers.Add("Access-Control-Expose-Headers", a.exposes)
	}
	if len(a.credentials) != 0 {
		headers.Add("Access-Control-Allow-Credentials", a.credentials)
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		for key, values := range headers {
			w.Header()[key] = values
		}

		if next != nil {
			next.ServeHTTP(w, r)
		}
	})
}
