// Package debug is a dummy package that is not yet implemented.
package debug

// SetMaxStack sets the maximum amount of memory that can be used by a single
// goroutine stack.
//
// Not implemented.
func SetMaxStack(n int) int {
	return n
}

// Stack returns a formatted stack trace of the goroutine that calls it.
//
// Not implemented.
func Stack() []byte {
	return nil
}
