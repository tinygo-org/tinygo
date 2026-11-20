//go:build nintendoswitch || (wasip1 && !wasip3)

package syscall

import "unsafe"

func Environ() []string {
	return environFromPointer(libc_environ)
}

//go:extern environ
var libc_environ *unsafe.Pointer
