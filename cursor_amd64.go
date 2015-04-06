package pubsub

import (
	"fmt"
	"math"
	"sync/atomic"
	"unsafe"
)

var wordSize = int(unsafe.Sizeof(uintptr(0)))

const cacheLineSize = 8

func init() {
	if wordSize != 8 {
		panic(fmt.Sprintf("expected word size of 8, not %d", wordSize))
	}
}

type cursor [cacheLineSize]int64

func newCursor(val, mask int) *cursor {
	a := [cacheLineSize]int64{int64(val), int64(mask)}
	c := cursor(a)
	return &c
}

func (c *cursor) next() int {
	return int((atomic.LoadInt64(&c[0]) + 1) & c[1])
}

func (c *cursor) pos() int {
	return int(atomic.LoadInt64(&c[0]))
}

func (c *cursor) inc() int {
	for {
		v1 := atomic.LoadInt64(&c[0])
		v2 := (v1 + 1) & c[1]

		if atomic.CompareAndSwapInt64(&c[0], v1, v2) {

			return int(v2)
		}
	}
}

func (c *cursor) reset() {
	atomic.StoreInt64(&c[0], int64(-1))
}

type cursorSlice []*cursor

func newCursorSlice(size, mask int) cursorSlice {
	cs := make(cursorSlice, size)
	for i := range cs {
		cs[i] = newCursor(-1, mask)
	}
	return cs
}

func (s cursorSlice) get(val int) *cursor {
	v1 := int64(-1)
	v2 := int64(val)
	for {
		for i := range s {
			if atomic.CompareAndSwapInt64(&s[i][0], v1, v2) {
				return s[i]
			}
		}
	}
}

func nextPow2(v int) int {
	return int(math.Pow(2, math.Ceil(math.Log2(float64(v)))))
}
