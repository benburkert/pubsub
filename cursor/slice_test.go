package cursor

import "testing"

func TestCursorSlice(t *testing.T) {
	cs := MakeSlice(10, 7)

	for i := range cs {
		c := cs[i]
		if p := c.Pos(); p != -1 {
			t.Fatalf("want initial pos(c)=%d, got %d", -1, p)
		}
	}

	for i := range cs {
		c := cs.Alloc(i)
		if p := c.Pos(); p != i {
			t.Fatalf("want alloc pos(c)=%d, got %d", i, p)
		}
	}
}
