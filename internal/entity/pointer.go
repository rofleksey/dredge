package entity

// ToPointer returns a pointer to v.
func ToPointer[T any](v T) *T {
	return &v
}

// FromPointer returns the value pointed to by p, or the zero value of T if p is nil.
func FromPointer[T any](p *T) T {
	if p == nil {
		var z T
		return z
	}
	return *p
}
