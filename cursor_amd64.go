package pubsub

import (
	"fmt"
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

type cursor *int64

func newCursor(val int) cursor {
	s := [cacheLineSize]int64{}
	s[0] = int64(val)
	return cursor(&s[0])
}

func next(c cursor, card int) int {
	return int((atomic.LoadInt64(c) + 1) % int64(card))
}

func pos(c cursor) int {
	return int(atomic.LoadInt64(c))
}

func inc(c cursor, card int) int {
	for {
		v1 := atomic.LoadInt64(c)
		v2 := (v1 + 1) % int64(card)

		if atomic.CompareAndSwapInt64(c, v1, v2) {
			return int(v2)
		}
	}

}

func reset(c cursor) {
	atomic.StoreInt64(c, -1)
}

type cursorSlice []cursor

func newCursorSlice(size int) cursorSlice {
	s := make([]int64, size*cacheLineSize)
	cs := make(cursorSlice, size)
	for i := range cs {
		cs[i] = cursor(&s[i*cacheLineSize])
		reset(cs[i])
	}
	return cs
}

func (s cursorSlice) get(val int) cursor {
	v1 := int64(-1)
	v2 := int64(val)
	for {
		for i := range s {
			if atomic.CompareAndSwapInt64(s[i], v1, v2) {
				return s[i]
			}
		}
	}
}
