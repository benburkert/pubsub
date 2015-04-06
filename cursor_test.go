package pubsub

import "testing"

func TestCursor(t *testing.T) {
	c := newCursor(5, 7)

	if p := c.pos(); p != 5 {
		t.Fatalf("want pos(c)=%d, got %d", 5, p)
	}

	c.inc()
	if p := c.pos(); p != 6 {
		t.Fatalf("want pos(c)=%d, got %d", 6, p)
	}

	c.inc()
	if p := c.pos(); p != 7 {
		t.Fatalf("want pos(c)=%d, got %d", 7, p)
	}

	c.inc()
	if p := c.pos(); p != 0 {
		t.Fatalf("cursor did not wrap: want pos(c)=0, got %d", p)
	}

	c.reset()
	if p := c.pos(); p != -1 {
		t.Fatalf("want pos(c)=%d, got %d", -1, p)
	}
}

func TestCursorSlice(t *testing.T) {
	cs := newCursorSlice(10, 7)

	for i := range cs {
		c := cs[i]
		if p := c.pos(); p != -1 {
			t.Fatalf("want initial pos(c)=%d, got %d", -1, p)
		}
	}

	for i := range cs {
		c := cs.get(i)
		if p := c.pos(); p != i {
			t.Fatalf("want get pos(c)=%d, got %d", i, p)
		}
	}
}
