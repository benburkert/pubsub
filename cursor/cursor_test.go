package cursor

import "testing"

func TestCursor(t *testing.T) {
	c := New(5, 7)

	if p := c.Pos(); p != 5 {
		t.Fatalf("want pos(c)=%d, got %d", 5, p)
	}

	c.Inc()
	if p := c.Pos(); p != 6 {
		t.Fatalf("want pos(c)=%d, got %d", 6, p)
	}

	c.Inc()
	if p := c.Pos(); p != 7 {
		t.Fatalf("want pos(c)=%d, got %d", 7, p)
	}

	c.Inc()
	if p := c.Pos(); p != 0 {
		t.Fatalf("cursor did not wrap: want pos(c)=0, got %d", p)
	}

	c.Reset()
	if p := c.Pos(); p != -1 {
		t.Fatalf("want pos(c)=%d, got %d", -1, p)
	}
}
