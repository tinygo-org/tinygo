//go:build !wasip1

package os

// pollFD is a literal empty struct on non-wasip1 targets. As long as the
// field is not the last in its containing struct, Go gives a zero-sized
// non-trailing field a true zero byte layout — file then occupies exactly
// the same space as it would without the field at all.
type pollFD struct{}

func (pollFD) Close() error { return nil }
func (pollFD) Exist() bool  { return false }
