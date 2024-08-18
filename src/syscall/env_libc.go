//go:build nintendoswitch || wasip1

package syscall

import (
	"unsafe"
)

func Environ() []string {

	// This function combines all the environment into a single allocation.
	// While this optimizes for memory usage and garbage collector
	// overhead, it does run the risk of potentially pinning a "large"
	// allocation if a user holds onto a single environment variable or
	// value.  Having each variable be its own allocation would make the
	// trade-off in the other direction.

	// calculate total memory required
	var length uintptr
	var vars int
	for environ := libc_environ; *environ != nil; {
		length += libc_strlen(*environ)
		vars++
		environ = (*unsafe.Pointer)(unsafe.Add(unsafe.Pointer(environ), unsafe.Sizeof(environ)))
	}

	// allocate our backing slice for the strings
	b := make([]byte, length)
	// and the slice we're going to return
	envs := make([]string, 0, vars)

	// loop over the environment again, this time copying over the data to the backing slice
	for environ := libc_environ; *environ != nil; {
		length = libc_strlen(*environ)
		// construct a Go string pointing at the libc-allocated environment variable data
		var envVar string
		rawEnvVar := (*struct {
			ptr    unsafe.Pointer
			length uintptr
		})(unsafe.Pointer(&envVar))
		rawEnvVar.ptr = *environ
		rawEnvVar.length = length
		// pull off the number of bytes we need for this environment variable
		var bs []byte
		bs, b = b[:length], b[length:]
		// copy over the bytes to the Go heap
		copy(bs, envVar)
		// convert trimmed slice to string
		s := *(*string)(unsafe.Pointer(&bs))
		// add s to our list of environment variables
		envs = append(envs, s)
		// environ++
		environ = (*unsafe.Pointer)(unsafe.Add(unsafe.Pointer(environ), unsafe.Sizeof(environ)))
	}
	return envs
}

//go:extern environ
var libc_environ *unsafe.Pointer
