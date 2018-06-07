package surface_test

import "sync"

// Buffers todo
type Buffers struct {
	p *sync.Pool
}

// NewBuffers is
func NewBuffers() *Buffers {
	return &Buffers{
		p: &sync.Pool{
			New: func() interface{} {
				b := make([]byte, 32*1024)
				return &b
			},
		},
	}
}

// Alloc buffer
func (b *Buffers) Alloc() interface{} {
	return b.p.Get()
}

// Free buffers
func (b *Buffers) Free(x interface{}) {
	b.p.Put(x)
}
