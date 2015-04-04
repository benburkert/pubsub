package pubsub

import "testing"

func TestCursor(t *testing.T) {
	card := 7
	c := newCursor(5)

	if p := pos(c); p != 5 {
		t.Fatalf("want pos(c)=%d, got %d", 5, p)
	}

	inc(c, card)
	if p := pos(c); p != 6 {
		t.Fatalf("want pos(c)=%d, got %d", 6, p)
	}

	inc(c, card)
	if p := pos(c); p != 0 {
		t.Fatalf("cursor did not wrap: want pos(c)=0, got %d", p)
	}

	reset(c)
	if p := pos(c); p != -1 {
		t.Fatalf("want pos(c)=%d, got %d", -1, p)
	}
}

func TestCursorSlice(t *testing.T) {
	cs := newCursorSlice(10)

	for i := range cs {
		c := cs[i]
		if p := pos(c); p != -1 {
			t.Fatalf("want initial pos(c)=%d, got %d", -1, p)
		}
	}

	for i := range cs {
		c := cs.get(i)
		if p := pos(c); p != i {
			t.Fatalf("want get pos(c)=%d, got %d", i, p)
		}
	}
}
