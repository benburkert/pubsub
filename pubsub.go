package pubsub

import (
	"errors"
	"math"
	"sync"
)

var (
	errClosed = errors.New("PubSub is closed")
	errMaxSub = errors.New("maxSubCount reached")
)

type markerChan chan struct{}

type PubSub struct {
	buffer *Buffer

	donec markerChan
	doneo sync.Once

	pubwg sync.WaitGroup
	subwg sync.WaitGroup

	submu            sync.Mutex
	subCount, subMax int
}

func New(minBufferSize, maxSubCount int) (*PubSub, error) {
	size := calcBufferSize(minBufferSize)
	if size < 2 {
		return nil, errors.New("minBufferSize must be > 1")
	}

	if maxSubCount < 1 {
		return nil, errors.New("maxSubCount must be > 0")
	}

	return &PubSub{
		buffer: NewBuffer(size, maxSubCount),
		donec:  make(markerChan),
		subMax: maxSubCount,
	}, nil
}

func (ps *PubSub) Close() {
	ps.doneo.Do(func() {
		ps.buffer.Write(ps.donec)
		close(ps.donec)
	})

	ps.pubwg.Wait()
	ps.subwg.Wait()
}

func (ps *PubSub) Pub(v interface{}) error {
	if ps.isClosed() {
		return errClosed
	}

	ps.buffer.Write(v)
	return nil
}

func (ps *PubSub) PubChan(ch <-chan interface{}) (<-chan struct{}, error) {
	if ps.isClosed() {
		return nil, errClosed
	}

	ps.pubwg.Add(1)
	go func() {
		defer ps.pubwg.Done()

		for {
			select {
			case v := <-ch:
				ps.buffer.Write(v)
			case <-ps.donec:
				for v := range ch {
					ps.buffer.Write(v)
				}
				return
			}
		}
	}()

	return ps.donec, nil
}

func (ps *PubSub) PubSlice(vs []interface{}) error {
	if ps.isClosed() {
		return errClosed
	}

	ps.buffer.WriteSlice(vs)
	return nil
}

func (ps *PubSub) SubChan(ch chan<- interface{}) (chan<- struct{}, error) {
	if ps.isClosed() {
		return nil, errClosed
	}
	if !ps.addSub() {
		return nil, errMaxSub
	}

	unsubc := make(markerChan)
	go func() {
		<-unsubc
		ps.Pub(unsubc)
	}()

	rfn := func(v interface{}) bool {
		if vch, ok := v.(markerChan); ok {
			if vch == ps.donec || vch == unsubc {
				close(ch)
				ps.delSub()
				return false
			}
		} else {
			ch <- v
		}
		return true
	}

	ps.buffer.ReadTo(rfn)
	return unsubc, nil
}

func (ps *PubSub) SubFunc(fn func(interface{})) (func(), error) {
	if ps.isClosed() {
		return nil, errClosed
	}
	if !ps.addSub() {
		return nil, errMaxSub
	}

	unsubc := make(markerChan)
	unsubfn := func() {
		ps.Pub(unsubc)
	}

	rfn := func(v interface{}) bool {
		if vch, ok := v.(markerChan); ok {
			if vch == ps.donec || vch == unsubc {
				ps.delSub()
				return false
			}
		} else {
			fn(v)
		}
		return true
	}

	ps.buffer.ReadTo(rfn)
	return unsubfn, nil
}

func (ps *PubSub) addSub() bool {
	ps.submu.Lock()
	defer ps.submu.Unlock()

	if ps.subCount == ps.subMax {
		return false
	}

	ps.subCount++
	ps.subwg.Add(1)
	return true
}

func (ps *PubSub) delSub() {
	ps.submu.Lock()
	defer ps.submu.Unlock()

	ps.subCount--
	ps.subwg.Done()
}

func (ps *PubSub) isClosed() bool {
	select {
	case <-ps.donec:
		return true
	default:
		return false
	}
}

func calcBufferSize(minSize int) int {
	return int(math.Ceil(math.Log2(float64(minSize + 1))))
}
