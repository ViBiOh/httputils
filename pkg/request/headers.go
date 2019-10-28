package request

import (
	"encoding/base64"
	"fmt"
)

// GenerateBasicAuth generates Basic Auth for given username and password
func GenerateBasicAuth(username string, password string) string {
	return fmt.Sprintf("Basic %s", base64.StdEncoding.EncodeToString([]byte(username+":"+password)))
}
