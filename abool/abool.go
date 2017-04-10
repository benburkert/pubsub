package abool

import "sync/atomic"

// Value provides an atomic boolean value.
type Value interface {
	// Set sets the boolean value to true.
	Set()
	// Test returns the boolean value.
	Test() bool
	// Unset sets the boolean value to false.
	Unset()
}

type value atomic.Value

// New returns a new Value initialized to b.
func New(b bool) Value {
	v := new(value)
	if b {
		v.Set()
	} else {
		v.Unset()
	}

	return v
}

func (v *value) Set() {
	(*atomic.Value)(v).Store(uint8(1))
}

func (v *value) Test() bool {
	return (*atomic.Value)(v).Load().(uint8) == 1
}

func (v *value) Unset() {
	(*atomic.Value)(v).Store(uint8(0))
}
