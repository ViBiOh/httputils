package model

import (
	"bytes"
	"sync"
)

var (
	// BufferPool for getting bytes buffer
	BufferPool = sync.Pool{
		New: func() interface{} {
			return bytes.NewBuffer(make([]byte, 1024))
		},
	}
)
