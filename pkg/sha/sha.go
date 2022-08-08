package sha

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"hash"
)

// New get sha256 value of given interface
func New[T any](o T) string {
	hasher := sha256.New()

	// no err check https://golang.org/pkg/hash/#Hash
	_, _ = fmt.Fprintf(hasher, "%#v", o)

	return hex.EncodeToString(hasher.Sum(nil))
}

// StreamHasher is a hasher encapsulation
type StreamHasher struct {
	hasher hash.Hash
}

// Stream create a new stream hasher
func Stream() StreamHasher {
	return StreamHasher{
		hasher: sha256.New(),
	}
}

// Write writes content to the hasher
func (s StreamHasher) Write(o any) StreamHasher {
	// no err check https://golang.org/pkg/hash/#Hash
	_, _ = fmt.Fprintf(s.hasher, "%#v", o)

	return s
}

// Sum returns the result of hashing
func (s StreamHasher) Sum() string {
	return hex.EncodeToString(s.hasher.Sum(nil))
}
