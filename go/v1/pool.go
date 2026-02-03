package v1

import "sync"

const (
	// DefaultBufferSize is the default size for buffers in the pool
	// 32KB should be enough for most control frames and small data chunks
	DefaultBufferSize = 32 * 1024
)

// BufferPool manages a pool of byte slices to reduce GC pressure
type BufferPool struct {
	pool sync.Pool
}

// NewBufferPool creates a new buffer pool
func NewBufferPool() *BufferPool {
	return &BufferPool{
		pool: sync.Pool{
			New: func() interface{} {
				// Allocate a new buffer of default size
				return make([]byte, DefaultBufferSize)
			},
		},
	}
}

// Get returns a buffer of at least the requested capacity
// If size is 0 or > DefaultBufferSize, it might allocate a new one or return a standard one.
// For simplicity: mostly fixed size buffers.
func (p *BufferPool) Get(size int) []byte {
	if size > DefaultBufferSize {
		return make([]byte, size)
	}

	v := p.pool.Get()
	if v == nil {
		return make([]byte, DefaultBufferSize)
	}

	buf := v.([]byte)
	if cap(buf) < size {
		// Should not happen if size <= DefaultBufferSize
		return make([]byte, size)
	}

	return buf[:size]
}

// Put returns a buffer to the pool
func (p *BufferPool) Put(buf []byte) {
	// Only pool buffers of standard size/cap to ensure consistency
	if cap(buf) != DefaultBufferSize {
		return
	}
	p.pool.Put(buf)
}

var globalPool = NewBufferPool()

// GetBuffer gets a buffer from global pool
func GetBuffer(size int) []byte {
	return globalPool.Get(size)
}

// PutBuffer puts a buffer back to global pool
func PutBuffer(buf []byte) {
	globalPool.Put(buf)
}
