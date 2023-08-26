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
	Origin      string
	Headers     string
	Methods     string
	Exposes     string
	Credentials bool
}

func Flags(fs *flag.FlagSet, prefix string, overrides ...flags.Override) Config {
	var config Config

	flags.New("Origin", "Access-Control-Allow-Origin").Prefix(prefix).DocPrefix("cors").StringVar(fs, &config.Origin, "*", overrides)
	flags.New("Headers", "Access-Control-Allow-Headers").Prefix(prefix).DocPrefix("cors").StringVar(fs, &config.Headers, "Content-Type", overrides)
	flags.New("Methods", "Access-Control-Allow-Methods").Prefix(prefix).DocPrefix("cors").StringVar(fs, &config.Methods, http.MethodGet, overrides)
	flags.New("Expose", "Access-Control-Expose-Headers").Prefix(prefix).DocPrefix("cors").StringVar(fs, &config.Exposes, "", overrides)
	flags.New("Credentials", "Access-Control-Allow-Credentials").Prefix(prefix).DocPrefix("cors").BoolVar(fs, &config.Credentials, false, overrides)

	return config
}

func New(config Config) App {
	return App{
		origin:      config.Origin,
		headers:     config.Headers,
		methods:     config.Methods,
		exposes:     config.Exposes,
		credentials: strconv.FormatBool(config.Credentials),
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
