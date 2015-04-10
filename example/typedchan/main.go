package main

import (
	"fmt"
	"sync"

	"github.com/benburkert/pubsub"
)

type foo struct {
	N int
}

type fooChan chan foo

func (ch fooChan) PublishTo(ctx *pubsub.Context) error {
	go func() {
		defer ctx.Close()

		for {
			select {
			case v, ok := <-ch:
				if !ok {
					return
				}

				ctx.Buffer.Write(v)
			case <-ctx.Done:
				return
			}
		}
	}()

	return nil
}

func (ch fooChan) SubscribeTo(ctx *pubsub.Context) error {
	rfn := func(v interface{}) bool {
		if v == ctx.Done {
			close(ch)
			ctx.Close()
			return false
		}
		ch <- v.(foo)
		return true
	}

	ctx.Buffer.ReadTo(rfn)
	return nil
}

func main() {
	ps, _ := pubsub.New(16, 4)

	pch := make(fooChan)

	if err := ps.AddPublisher(pch); err != nil {
		panic(err)
	}

	wg := &sync.WaitGroup{}
	wg.Add(4)

	for _, id := range []string{"A", "B", "C", "D"} {
		ch := make(fooChan, 4)
		ps.AddSubscriber(ch)

		go func(ch <-chan foo, id string) {
			defer wg.Done()

			for v := range ch {
				fmt.Printf("%s got foo.N=%d\n", id, v.N)
			}
		}(ch, id)
	}

	for i := 0; i <= 25; i++ {
		pch <- foo{N: i}
	}
	ps.Close()

	wg.Wait()
}
