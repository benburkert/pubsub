package main

import (
	"fmt"

	"github.com/benburkert/pubsub"
)

func main() {
	ps, _ := pubsub.New(16, 4)

	subs := []string{"A", "B", "C", "D"}
	for i := range subs {
		id := subs[i]
		fn := func(v interface{}) {
			fmt.Printf("%s got %d\n", id, v)
		}
		ps.SubFunc(fn)
	}

	for i := 0; i <= 25; i++ {
		ps.Pub(i)
	}
	ps.Close()
}
