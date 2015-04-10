package main

import (
	"fmt"

	"github.com/benburkert/pubsub"
)

type foo struct {
	N int
}

type fooFunc func(v foo)

func (fn fooFunc) SubscribeTo(ctx *pubsub.Context) error {
	rfn := func(v interface{}) bool {
		if v == ctx.Done {
			ctx.Close()
			return false
		}

		fn(v.(foo))
		return true
	}

	ctx.Buffer.ReadTo(rfn)
	return nil
}

func main() {
	ps, _ := pubsub.New(16, 4)

	subs := []string{"A", "B", "C", "D"}
	for i := range subs {
		id := subs[i]
		fn := fooFunc(func(v foo) {
			fmt.Printf("%s got foo.N=%d\n", id, v.N)
		})
		ps.AddSubscriber(fn)
	}

	for i := 0; i <= 25; i++ {
		ps.Pub(foo{N: i})
	}
	ps.Close()
}
