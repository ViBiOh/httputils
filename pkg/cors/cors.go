package cors

import (
	"flag"
	"maps"
	"net/http"
	"strconv"

	"github.com/ViBiOh/flags"
	"github.com/ViBiOh/httputils/v4/pkg/model"
)

var _ model.Middleware = Service{}.Middleware

type Service struct {
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

func Flags(fs *flag.FlagSet, prefix string, overrides ...flags.Override) *Config {
	var config Config

	flags.New("Origin", "Access-Control-Allow-Origin").Prefix(prefix).DocPrefix("cors").StringVar(fs, &config.Origin, "*", overrides)
	flags.New("Headers", "Access-Control-Allow-Headers").Prefix(prefix).DocPrefix("cors").StringVar(fs, &config.Headers, "Content-Type", overrides)
	flags.New("Methods", "Access-Control-Allow-Methods").Prefix(prefix).DocPrefix("cors").StringVar(fs, &config.Methods, http.MethodGet, overrides)
	flags.New("Expose", "Access-Control-Expose-Headers").Prefix(prefix).DocPrefix("cors").StringVar(fs, &config.Exposes, "", overrides)
	flags.New("Credentials", "Access-Control-Allow-Credentials").Prefix(prefix).DocPrefix("cors").BoolVar(fs, &config.Credentials, false, overrides)

	return &config
}

func New(config *Config) Service {
	return Service{
		origin:      config.Origin,
		headers:     config.Headers,
		methods:     config.Methods,
		exposes:     config.Exposes,
		credentials: strconv.FormatBool(config.Credentials),
	}
}

func (s Service) Middleware(next http.Handler) http.Handler {
	headers := http.Header{}

	if len(s.origin) != 0 {
		headers.Add("Access-Control-Allow-Origin", s.origin)
	}
	if len(s.headers) != 0 {
		headers.Add("Access-Control-Allow-Headers", s.headers)
	}
	if len(s.methods) != 0 {
		headers.Add("Access-Control-Allow-Methods", s.methods)
	}
	if len(s.exposes) != 0 {
		headers.Add("Access-Control-Expose-Headers", s.exposes)
	}
	if len(s.credentials) != 0 {
		headers.Add("Access-Control-Allow-Credentials", s.credentials)
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		maps.Copy(w.Header(), headers)

		if next != nil {
			next.ServeHTTP(w, r)
		}
	})
}
