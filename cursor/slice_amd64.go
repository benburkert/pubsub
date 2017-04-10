package cursor

import "sync/atomic"

// Alloc allocates the next unused or reset Cursor.
func (s Slice) Alloc(pos int) *Cursor {
	for {
		for i := range s {
			if atomic.CompareAndSwapInt64(&s[i].pos, -1, int64(pos)) {
				return s[i]
			}
		}
	}
}
