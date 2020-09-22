// +build !go1.14

package runtime

// Runtime timer, must be kept in sync with src/time/sleep.go of the Go stdlib.
// The layout changed in Go 1.14, so this is only supported on Go 1.13 and
// below.
//
// The fields used by the time package are:
//   when:    time to wake up (in nanoseconds)
//   period:  if not 0, a repeating time (of the given nanoseconds)
//   f:       the function called when the timer expires
//   arg:     parameter to f
type timer struct {
	tb uintptr
	i  int

	when   int64
	period int64
	f      func(interface{}, uintptr)
	arg    interface{}
	seq    uintptr
}
