package hash

import (
	"encoding/hex"
	"fmt"
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
	hasher *xxh3.Hasher
}

func Stream() StreamHasher {
	return StreamHasher{
		hasher: xxh3.New(),
	}
}

func (s StreamHasher) WriteString(o string) StreamHasher {
	_, _ = s.hasher.WriteString(o)

	return s
}

func (s StreamHasher) WriteBytes(o []byte) StreamHasher {
	_, _ = s.hasher.Write(o)

	return s
}

func (s StreamHasher) Write(o any) StreamHasher {
	switch typedValue := o.(type) {
	case []byte:
		_, _ = s.hasher.Write(typedValue)

	case string:
		_, _ = s.hasher.WriteString(typedValue)

	case int:
		_, _ = s.hasher.WriteString(strconv.Itoa(typedValue))

	case int64:
		_, _ = s.hasher.WriteString(strconv.FormatInt(typedValue, 10))

	case uint64:
		_, _ = s.hasher.WriteString(strconv.FormatUint(typedValue, 10))

	case float64:
		_, _ = s.hasher.WriteString(strconv.FormatFloat(typedValue, 'f', -1, 64))

	case bool:
		if typedValue {
			_, _ = s.hasher.WriteString("true")
		} else {
			_, _ = s.hasher.WriteString("false")
		}

	default:
		var buffer [64]byte
		_, _ = s.hasher.Write(strconv.AppendQuote(buffer[:0], fmt.Sprintf("%v", typedValue)))
	}

	return s
}

func (s StreamHasher) Sum() string {
	return hex.EncodeToString(s.hasher.Sum(nil))
}
