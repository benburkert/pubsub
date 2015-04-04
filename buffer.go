package pubsub

import "sync"

const (
	empty = iota
)

type ReaderFunc func(interface{}) bool

type Buffer struct {
	mu   sync.RWMutex
	cond *sync.Cond
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
		b.data[i] = empty
	}

	b.wcond = sync.NewCond(&b.mu)
	b.rcond = sync.NewCond(b.mu.RLocker())
	return b
}

func (b *Buffer) Read() []interface{} {
	b.mu.RLock()
	defer b.mu.RUnlock()

	cursor := b.getCursor()
	defer reset(cursor)

	rpos := pos(cursor)
	if b.data[rpos] == empty {
		s := make([]interface{}, rpos)
		copy(s, b.data[:rpos])
		return s
	}

	s := make([]interface{}, b.card)
	copy(s[:(b.card-rpos)], b.data[rpos:])
	copy(s[(b.card-rpos):], b.data[:rpos])
	return s
}

func (b *Buffer) ReadTo(rfn ReaderFunc) {
	b.mu.RLock() // unlocked in readTo

	go b.readTo(b.getCursor(), rfn)
}

func (b *Buffer) Write(v interface{}) {
	b.mu.Lock()
	defer b.mu.Unlock()

	for b.writeBarrier() {
		b.wcond.Wait()
	}

	wpos := pos(b.wcursor)
	b.data[wpos] = v
	inc(b.wcursor, b.card)

	b.rcond.Broadcast()
}

// assumes b.mu RLock held
func (b *Buffer) getCursor() cursor {
	return b.rcursors.get(pos(b.wcursor))
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
