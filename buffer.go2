package pubsub

import (
	"math"
	"sync"

	"github.com/benburkert/pubsub/cursor"
)

type Marker byte

const (
	EmptyMarker Marker = iota
)

type ReaderFunc func(interface{}) bool

type Buffer struct {
	mu   sync.RWMutex
	data []interface{}

	wcond   *sync.Cond
	wcursor *cursor.Cursor

	rcond    *sync.Cond
	rcursors cursor.Slice
}

func NewBuffer(minSize, maxReaders int) *Buffer {
	size := calcBufferSize(minSize)
	mask := size - 1

	b := &Buffer{
		data:     make([]interface{}, size),
		wcursor:  cursor.New(0, mask),
		rcursors: cursor.MakeSlice(maxReaders, mask),
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
	defer c.Reset()

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
func (b *Buffer) getCursor() *cursor.Cursor {
	return b.rcursors.Alloc(b.wcursor.Pos())
}

// assumes b.mu Rlock held
func (b *Buffer) read(c *cursor.Cursor) []interface{} {
	rpos := c.Pos()
	if b.data[rpos] == EmptyMarker {
		s := make([]interface{}, rpos)
		copy(s, b.data[:rpos])
		return s
	}

	size := int(len(b.data))
	s := make([]interface{}, size)
	copy(s[:(size-rpos)], b.data[rpos:])
	copy(s[(size-rpos):], b.data[:rpos])
	return s
}

// assumes b.mu RLock held
func (b *Buffer) readBarrier(c *cursor.Cursor) bool {
	return c.Pos() == b.wcursor.Pos()
}

// asumes b.mu RLock held
func (b *Buffer) readTo(c *cursor.Cursor, rfn ReaderFunc) {
	defer b.mu.RUnlock()
	defer b.wcond.Signal()
	defer c.Reset()

	for {
		for b.readBarrier(c) {
			b.rcond.Wait()
		}

		rpos, wpos := c.Pos(), b.wcursor.Pos()
		for rpos != wpos {
			if !rfn(b.data[rpos]) {
				return
			}
			rpos = c.Inc()
		}
		b.wcond.Signal()
	}
}

// asumes b.mu Lock held
func (b *Buffer) write(v interface{}) {
	for b.writeBarrier() {
		b.wcond.Wait()
	}

	wpos := b.wcursor.Pos()
	b.data[wpos] = v
	b.wcursor.Inc()

	b.rcond.Broadcast()
}

// assumes b.mu Lock held
func (b *Buffer) writeBarrier() bool {
	npos := b.wcursor.Next()
	for _, c := range b.rcursors {
		if npos == c.Pos() {
			return true
		}
	}
	return false
}

func calcBufferSize(minSize int) int {
	return int(math.Pow(2, math.Ceil(math.Log2(float64(minSize)))))
}
