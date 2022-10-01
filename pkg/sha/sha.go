package sha

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"hash"
)

func New(content any) string {
	hasher := sha256.New()

	// no err check https://golang.org/pkg/hash/#Hash
	_, _ = fmt.Fprintf(hasher, "%#v", content)

	return hex.EncodeToString(hasher.Sum(nil))
}

type StreamHasher struct {
	hasher hash.Hash
}

func Stream() StreamHasher {
	return StreamHasher{
		hasher: sha256.New(),
	}
}

func (s StreamHasher) Write(o any) StreamHasher {
	// no err check https://golang.org/pkg/hash/#Hash
	_, _ = fmt.Fprintf(s.hasher, "%#v", o)

	return s
}

func (s StreamHasher) Sum() string {
	return hex.EncodeToString(s.hasher.Sum(nil))
}
