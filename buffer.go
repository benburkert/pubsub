package pubsub

import "sync"

type Marker int

const (
	EmptyMarker Marker = iota
)

type ReaderFunc func(interface{}) bool

type Buffer struct {
	mu   sync.RWMutex
	data []interface{}
	card int

	wcond   *sync.Cond
	wcursor cursor

	rcond    *sync.Cond
	rcursors cursorSlice
}

func NewBuffer(bufLen, maxReaders int) *Buffer {
	b := &Buffer{
		data:     make([]interface{}, bufLen),
		card:     bufLen,
		wcursor:  newCursor(0),
		rcursors: newCursorSlice(maxReaders),
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
	defer reset(c)

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
func (b *Buffer) getCursor() cursor {
	return b.rcursors.get(pos(b.wcursor))
}

// assumes b.mu Rlock held
func (b *Buffer) read(c cursor) []interface{} {
	rpos := pos(c)
	if b.data[rpos] == EmptyMarker {
		s := make([]interface{}, rpos)
		copy(s, b.data[:rpos])
		return s
	}

	s := make([]interface{}, b.card)
	copy(s[:(b.card-rpos)], b.data[rpos:])
	copy(s[(b.card-rpos):], b.data[:rpos])
	return s
}

// assumes b.mu RLock held
func (b *Buffer) readBarrier(c cursor) bool {
	return pos(c) == pos(b.wcursor)
}

// asumes b.mu RLock held
func (b *Buffer) readTo(c cursor, rfn ReaderFunc) {
	defer b.mu.RUnlock()
	defer b.wcond.Signal()
	defer reset(c)

	for {
		for b.readBarrier(c) {
			b.rcond.Wait()
		}

		rpos, wpos := pos(c), pos(b.wcursor)
		for rpos != wpos {
			if !rfn(b.data[rpos]) {
				return
			}
			rpos = inc(c, b.card)
		}
		b.wcond.Signal()
	}
}

// asumes b.mu Lock held
func (b *Buffer) write(v interface{}) {
	for b.writeBarrier() {
		b.wcond.Wait()
	}

	wpos := pos(b.wcursor)
	b.data[wpos] = v
	inc(b.wcursor, b.card)

	b.rcond.Broadcast()
}

// assumes b.mu Lock held
func (b *Buffer) writeBarrier() bool {
	npos := next(b.wcursor, b.card)
	for _, c := range b.rcursors {
		if npos == pos(c) {
			return true
		}
	}
	return false
}
