//go:build avr

package reflect

// intw is an integer type, used in places where an int is typically required,
// except architectures where the size of an int != word size.
// See https://github.com/tinygo-org/tinygo/issues/1284.
type intw = uintptr
