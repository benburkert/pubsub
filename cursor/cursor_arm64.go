package cursor

import "sync/atomic"

// Cursor marks a position in a ring buffer.
type Cursor struct {
	pos, mask int64

	pad [48]byte
}

// New allocates a new Cursor at pos for a ring buffer mask.
func New(pos, mask int) *Cursor {
	return &Cursor{
		pos:  int64(pos),
		mask: int64(mask),
	}
}

// Next returns the next position index.
func (c *Cursor) Next() int {
	return int((atomic.LoadInt64(&c.pos) + 1) & c.mask)
}

// Pos returns the current position index.
func (c *Cursor) Pos() int {
	return int(atomic.LoadInt64(&c.pos))
}

// Inc moves the position forward one space.
func (c *Cursor) Inc() int {
	for {
		v1 := atomic.LoadInt64(&c.pos)
		v2 := (v1 + 1) & c.mask

		if atomic.CompareAndSwapInt64(&c.pos, v1, v2) {
			return int(v2)
		}
	}
}

// Reset clears the position.
func (c *Cursor) Reset() {
	atomic.StoreInt64(&c.pos, -1)
}
