//go:build wasip2

package syscall

import "unsafe"

func Environ() []string {
	// The environ symbol is filled by a constructor and traps in cabi_realloc.
	// Use the lazy accessor instead.
	return environFromPointer(libc_get_environ())
}

//export __wasilibc_get_environ
func libc_get_environ() *unsafe.Pointer
