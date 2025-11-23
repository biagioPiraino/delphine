package types

import (
	"bytes"
	"sync"
)

var JsonBufferPool = sync.Pool{
	New: func() any {
		return bytes.NewBuffer(make([]byte, 0, 4096)) // preallocate 4kb of bytes
	},
}
