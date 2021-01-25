package pubsub

import (
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
)

func TestPubSubErrors(t *testing.T) {
	if _, err := New(1, 1); err == nil {
		t.Error("expected error for minBufferSize=1")
	}
	if _, err := New(2, 0); err == nil {
		t.Error("expected error for maxSubCount=0")
	}

	ps, err := New(2, 1)
	if err != nil {
		t.Fatal(err)
	}

	if _, err = ps.SubChan(make(chan interface{})); err != nil {
		t.Fatal(err)
	}
	if _, err = ps.SubChan(make(chan interface{})); err != errMaxSub {
		t.Errorf("unexpected error %q", err)
	}

	ps.Close()
	if _, err = ps.SubChan(make(chan interface{})); err != errClosed {
		t.Errorf("unexpected error %q", err)
	}
}

func TestPubSubFuncs(t *testing.T) {
	ps, err := New(4, 2)
	if err != nil {
		t.Fatal(err)
	}
	defer ps.Close()

	counter := new(int32)
	fn := func(interface{}) { atomic.AddInt32(counter, 1) }

	if _, err := ps.SubFunc(fn); err != nil {
		t.Fatal(err)
	}
	if _, err := ps.SubFunc(fn); err != nil {
		t.Fatal(err)
	}

	for i := 0; i < 16; i++ {
		ps.Pub(struct{}{})
	}

	if count := atomic.LoadInt32(counter); count != 32 {
		fmt.Errorf("want count=32, got %d", count)
	}
}

func TestPubSubChans(t *testing.T) {
	ps, err := New(4, 2)
	if err != nil {
		t.Fatal(err)
	}
	defer ps.Close()

	wg := &sync.WaitGroup{}
	wg.Add(32)

	for i := 0; i < 2; i++ {
		ch := make(chan interface{})
		go func(ch <-chan interface{}) {
			for range ch {
				wg.Done()
			}
		}(ch)
		if _, err := ps.SubChan(ch); err != nil {
			t.Fatal(err)
		}
	}

	for i := 0; i < 16; i++ {
		ps.Pub(struct{}{})
	}

	wg.Wait()
}

func TesPubSubMixed(t *testing.T) {
	ps, err := New(4, 2)
	if err != nil {
		t.Fatal(err)
	}
	defer ps.Close()

	counter := new(int32)
	fn := func(interface{}) { atomic.AddInt32(counter, 1) }

	if _, err := ps.SubFunc(fn); err != nil {
		t.Fatal(err)
	}

	wg := &sync.WaitGroup{}
	wg.Add(32)
	ch := make(chan interface{})
	go func() {
		for range ch {
			fn(struct{}{})
		}
	}()

	if _, err := ps.SubChan(ch); err != nil {
		t.Fatal(err)
	}

	for i := 0; i < 16; i++ {
		ps.Pub(struct{}{})
	}

	wg.Wait()
	if count := atomic.LoadInt32(counter); count != 32 {
		fmt.Errorf("want count=32, got %d", count)
	}
}

func TestPubSubFuncUnsubscribe(t *testing.T) {
	ps, err := New(4, 2)
	if err != nil {
		t.Fatal(err)
	}
	defer ps.Close()

	var unsubfn func()
	count := 0
	stepper := make(chan struct{})
	subfn := func(interface{}) {
		if count == 5 {
			go func() {
				unsubfn()
				close(stepper)
			}()

			return
		}
		if count > 5 {
			t.Error("value recieved after unsubscribing")
			return
		}
		stepper <- struct{}{}
		count++
	}

	if unsubfn, err = ps.SubFunc(subfn); err != nil {
		t.Fatal(err)
	}

	go func() { stepper <- struct{}{} }()
	for i := 0; i <= 10; i++ {
		<-stepper
		ps.Pub(struct{}{})
	}
}

func TestPubSubChanUnsubscribe(t *testing.T) {
	ps, err := New(4, 2)
	if err != nil {
		t.Fatal(err)
	}
	defer ps.Close()

	ch := make(chan interface{})
	unsubch, err := ps.SubChan(ch)
	if err != nil {
		t.Fatal(err)
	}

	donec := make(chan struct{})
	go func() {
		count := 0
		for range ch {
			if count == 5 {
				close(unsubch)
			}
			count++
		}
		close(donec)
	}()

	for i := 0; i < 10; i++ {
		ps.Pub(struct{}{})
	}
	<-donec
}
