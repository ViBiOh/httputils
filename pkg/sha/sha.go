package sha

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
)

// New get sha1 value of given interface
func New(o interface{}) string {
	hasher := sha1.New()

	// no err check https://golang.org/pkg/hash/#Hash
	_, _ = fmt.Fprintf(hasher, "%#v", o)

	return hex.EncodeToString(hasher.Sum(nil))
}
