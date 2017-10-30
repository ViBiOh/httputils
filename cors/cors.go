package cors

import (
	"flag"
	"net/http"
	"strconv"

	"github.com/ViBiOh/httputils/tools"
)

const defaultOrigin = `*`
const defaultHeaders = `Content-Type`
const defaultMethods = http.MethodGet
const defaultExposes = ``
const defaultCredentials = false

// Flags add flags for given prefix
func Flags(prefix string) map[string]interface{} {
	return map[string]interface{}{
		`origin`:      flag.String(tools.ToCamel(prefix+`Origin`), defaultOrigin, `[cors] Access-Control-Allow-Origin`),
		`headers`:     flag.String(tools.ToCamel(prefix+`Headers`), defaultHeaders, `[cors] Access-Control-Allow-Headers`),
		`methods`:     flag.String(tools.ToCamel(prefix+`Methods`), defaultMethods, `[cors] Access-Control-Allow-Methods`),
		`exposes`:     flag.String(tools.ToCamel(prefix+`Expose`), defaultExposes, `[cors] Access-Control-Expose-Headers`),
		`credentials`: flag.Bool(tools.ToCamel(prefix+`Credentials`), defaultCredentials, `[cors] Access-Control-Allow-Credentials`),
	}
}

// Handler for net/http package allowing cors header
func Handler(config map[string]interface{}, next http.Handler) http.Handler {
	var (
		origin      = defaultOrigin
		headers     = defaultHeaders
		methods     = defaultMethods
		exposes     = defaultExposes
		credentials = defaultCredentials
	)

	var given interface{}
	var ok bool

	if given, ok = config[`origin`]; ok {
		origin = *(given.(*string))
	}
	if given, ok = config[`headers`]; ok {
		headers = *(given.(*string))
	}
	if given, ok = config[`methods`]; ok {
		methods = *(given.(*string))
	}
	if given, ok = config[`exposes`]; ok {
		exposes = *(given.(*string))
	}
	if given, ok = config[`credentials`]; ok {
		credentials = *(given.(*bool))
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add(`Access-Control-Allow-Origin`, origin)
		w.Header().Add(`Access-Control-Allow-Headers`, headers)
		w.Header().Add(`Access-Control-Allow-Methods`, methods)
		w.Header().Add(`Access-Control-Expose-Headers`, exposes)
		w.Header().Add(`Access-Control-Allow-Credentials`, strconv.FormatBool(credentials))

		next.ServeHTTP(w, r)
	})
}
