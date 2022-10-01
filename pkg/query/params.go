package query

import (
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

func GetBool(r *http.Request, name string) bool {
	params, err := url.ParseQuery(r.URL.RawQuery)
	if err != nil {
		return false
	}

	value, ok := params[name]
	if !ok {
		return false
	}

	strValue := strings.Join(value, "")
	if len(strValue) == 0 {
		return true
	}

	boolValue, err := strconv.ParseBool(strValue)
	if err != nil {
		return false
	}

	return boolValue
}
