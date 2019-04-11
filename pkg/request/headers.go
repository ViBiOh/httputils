package request

import (
	"encoding/base64"
	"fmt"
	"net/http"
)

const (
	// ForwardedForHeader that proxy uses to fill
	ForwardedForHeader = "X-Forwarded-For"
)

// GenerateBasicAuth generates Basic Auth for given username and password
func GenerateBasicAuth(username string, password string) string {
	return fmt.Sprintf("Basic %s", base64.StdEncoding.EncodeToString([]byte(username+":"+password)))
}

// SetIP set remote IP
func SetIP(r *http.Request, ip string) {
	r.Header.Set(ForwardedForHeader, ip)
}

// GetIP give remote IP
func GetIP(r *http.Request) (ip string) {
	ip = r.Header.Get(ForwardedForHeader)
	if ip == "" {
		ip = r.RemoteAddr
	}

	return
}
