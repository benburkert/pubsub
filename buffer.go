package pubsub

import (
	"math"
	"sync"
)

type Marker int

const (
	EmptyMarker Marker = iota
)

type ReaderFunc func(interface{}) bool

type Buffer struct {
	mu   sync.RWMutex
	data []interface{}

	wcond   *sync.Cond
	wcursor *cursor

	rcond    *sync.Cond
	rcursors cursorSlice
}

func NewBuffer(minSize, maxReaders int) *Buffer {
	size := calcBufferSize(minSize)
	mask := size - 1

	b := &Buffer{
		data:     make([]interface{}, size),
		wcursor:  newCursor(0, mask),
		rcursors: newCursorSlice(maxReaders, mask),
	}

	for i := range b.data {
		b.data[i] = EmptyMarker
	}

	b.wcond = sync.NewCond(&b.mu)
	b.rcond = sync.NewCond(b.mu.RLocker())
	return b
}

func (b *Buffer) FullReadTo(rfn ReaderFunc) []interface{} {
	b.mu.RLock() // unlocked in readTo

	c := b.getCursor() // reset in readTo
	s := b.read(c)

	go b.readTo(c, rfn)
	return s
}

func (b *Buffer) Read() []interface{} {
	b.mu.RLock()
	defer b.mu.RUnlock()

	c := b.getCursor()
	defer c.reset()

	return b.read(c)
}

func (b *Buffer) ReadTo(rfn ReaderFunc) {
	b.mu.RLock() // unlocked in readTo

	go b.readTo(b.getCursor(), rfn)
}

func (b *Buffer) Write(v interface{}) {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.write(v)
}

func (b *Buffer) WriteSlice(vs []interface{}) {
	b.mu.Lock()
	defer b.mu.Unlock()

	for _, v := range vs {
		b.write(v)
	}
}

// assumes b.mu RLock held
func (b *Buffer) getCursor() *cursor {
	return b.rcursors.get(b.wcursor.pos())
}

// assumes b.mu Rlock held
func (b *Buffer) read(c *cursor) []interface{} {
	rpos := c.pos()
	if b.data[rpos] == EmptyMarker {
		s := make([]interface{}, rpos)
		copy(s, b.data[:rpos])
		return s
	}

	size := len(b.data)
	s := make([]interface{}, size)
	copy(s[:(size-rpos)], b.data[rpos:])
	copy(s[(size-rpos):], b.data[:rpos])
	return s
}

// assumes b.mu RLock held
func (b *Buffer) readBarrier(c *cursor) bool {
	return c.pos() == b.wcursor.pos()
}

// asumes b.mu RLock held
func (b *Buffer) readTo(c *cursor, rfn ReaderFunc) {
	defer b.mu.RUnlock()
	defer b.wcond.Signal()
	defer c.reset()

	for {
		for b.readBarrier(c) {
			b.rcond.Wait()
		}

		rpos, wpos := c.pos(), b.wcursor.pos()
		for rpos != wpos {
			if !rfn(b.data[rpos]) {
				return
			}
			rpos = c.inc()
		}
		b.wcond.Signal()
	}
}

// asumes b.mu Lock held
func (b *Buffer) write(v interface{}) {
	for b.writeBarrier() {
		b.wcond.Wait()
	}

	wpos := b.wcursor.pos()
	b.data[wpos] = v
	b.wcursor.inc()

	b.rcond.Broadcast()
}

// assumes b.mu Lock held
func (b *Buffer) writeBarrier() bool {
	npos := b.wcursor.next()
	for _, c := range b.rcursors {
		if npos == c.pos() {
			return true
		}
	}
	return false
}

func calcBufferSize(minSize int) int {
	return int(math.Pow(2, math.Ceil(math.Log2(float64(minSize)))))
}
