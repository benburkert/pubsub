package abool

import "testing"

func TestValue(t *testing.T) {
	if b := New(false); b.Test() {
		t.Fatal("false bool is true")
	}

	if b := New(true); !b.Test() {
		t.Fatal("true bool is false")
	}

	v := New(false)
	if v.Set(); !v.Test() {
		t.Fatal("set bool is false")
	}
	if v.Unset(); v.Test() {
		t.Fatal("unset bool is true")
	}
	if v.Set(); !v.Test() {
		t.Fatal("reset bool is false")
	}
}
