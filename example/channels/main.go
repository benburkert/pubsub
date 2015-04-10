package main

import (
	"fmt"
	"sync"

	"github.com/benburkert/pubsub"
)

func main() {
	ps, _ := pubsub.New(2, 4)

	wg := &sync.WaitGroup{}
	wg.Add(4)

	for _, id := range []string{"A", "B", "C", "D"} {
		ch := make(chan interface{}, 4)
		ps.SubChan(ch)

		go func(ch <-chan interface{}, id string) {
			defer wg.Done()

			for v := range ch {
				fmt.Printf("%s got %d\n", id, v)
			}
		}(ch, id)
	}

	ch := make(chan interface{})
	donec, _ := ps.PubChan(ch)
	for i := 0; i <= 25; i++ {
		select {
		case ch <- i:
		case <-donec:
			panic("PubSub closed unexpectedly")
		}
	}
	close(ch)
	ps.Close()
	wg.Wait()
}
