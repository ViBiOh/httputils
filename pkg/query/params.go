package query

import (
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

// GetBool converts query params to boolean
func GetBool(r *http.Request, name string) bool {
	if params, err := url.ParseQuery(r.URL.RawQuery); err == nil {
		if value, ok := params[name]; ok {
			strValue := strings.Join(value, ``)
			if strValue == `` {
				return true
			}

			pretty, err := strconv.ParseBool(strValue)
			if err != nil {
				return false
			}
			return pretty
		}
	}

	return false
}
