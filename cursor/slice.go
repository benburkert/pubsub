package cursor

// Slice is a slice of cursors
type Slice []*Cursor

// MakeSlice returns a new Slice of cursors with a common mask.
func MakeSlice(size, mask int) Slice {
	s := make(Slice, size)
	for i := range s {
		s[i] = New(-1, mask)
	}
	return s
}
