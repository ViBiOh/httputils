package hash

import (
	"encoding/hex"
	"fmt"
	"hash"
	"strconv"

	"github.com/zeebo/xxh3"
)

func String(value string) string {
	return strconv.FormatUint(xxh3.HashString(value), 16)
}

func Hash(content any) string {
	hasher := xxh3.New()

	_, _ = fmt.Fprintf(hasher, "%v", content)

	return hex.EncodeToString(hasher.Sum(nil))
}

type StreamHasher struct {
	hasher hash.Hash
}

func Stream() StreamHasher {
	return StreamHasher{
		hasher: xxh3.New(),
	}
}

func (s StreamHasher) WriteString(o string) StreamHasher {
	// no err check https://golang.org/pkg/hash/#Hash
	_, _ = s.hasher.Write([]byte(o))

	return s
}

func (s StreamHasher) WriteBytes(o []byte) StreamHasher {
	// no err check https://golang.org/pkg/hash/#Hash
	_, _ = s.hasher.Write(o)

	return s
}

func (s StreamHasher) Write(o any) StreamHasher {
	// no err check https://golang.org/pkg/hash/#Hash
	_, _ = fmt.Fprintf(s.hasher, "%v", o)

	return s
}

func (s StreamHasher) Sum() string {
	return hex.EncodeToString(s.hasher.Sum(nil))
}
