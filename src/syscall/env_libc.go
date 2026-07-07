//go:build nintendoswitch || wasip1 || wasip2

package syscall

import (
	"unsafe"
)

// environFromPointer walks a NULL-terminated char** environment array (in
// the same layout as libc's environ) and converts it into a []string. The
// starting pointer is supplied by the caller since how it's obtained
// (directly vs. via an accessor function) differs by platform/target -- see
// environ_libc.go and environ_wasip2.go.
func environFromPointer(environ *unsafe.Pointer) []string {
	// This function combines all the environment into a single allocation.
	// While this optimizes for memory usage and garbage collector
	// overhead, it does run the risk of potentially pinning a "large"
	// allocation if a user holds onto a single environment variable or
	// value.  Having each variable be its own allocation would make the
	// trade-off in the other direction.

	// calculate total memory required
	var length uintptr
	var vars int
	for e := environ; *e != nil; {
		length += libc_strlen(*e)
		vars++
		e = (*unsafe.Pointer)(unsafe.Add(unsafe.Pointer(e), unsafe.Sizeof(e)))
	}

	// allocate our backing slice for the strings
	b := make([]byte, length)
	// and the slice we're going to return
	envs := make([]string, 0, vars)

	// loop over the environment again, this time copying over the data to the backing slice
	for e := environ; *e != nil; {
		length = libc_strlen(*e)
		// construct a Go string pointing at the libc-allocated environment variable data
		var envVar string
		rawEnvVar := (*struct {
			ptr    unsafe.Pointer
			length uintptr
		})(unsafe.Pointer(&envVar))
		rawEnvVar.ptr = *e
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
		// e++
		e = (*unsafe.Pointer)(unsafe.Add(unsafe.Pointer(e), unsafe.Sizeof(e)))
	}
	return envs
}

func Getenv(key string) (value string, found bool) {
	data := cstring(key)
	raw := libc_getenv(&data[0])
	if raw == nil {
		return "", false
	}

	ptr := uintptr(unsafe.Pointer(raw))
	for size := uintptr(0); ; size++ {
		v := *(*byte)(unsafe.Pointer(ptr))
		if v == 0 {
			return string(unsafe.Slice(raw, size)), true
		}
		ptr += unsafe.Sizeof(byte(0))
	}
}

func Setenv(key, val string) (err error) {
	if len(key) == 0 {
		return EINVAL
	}
	for i := 0; i < len(key); i++ {
		if key[i] == '=' || key[i] == 0 {
			return EINVAL
		}
	}
	for i := 0; i < len(val); i++ {
		if val[i] == 0 {
			return EINVAL
		}
	}
	runtimeSetenv(key, val)
	return
}

func Unsetenv(key string) (err error) {
	runtimeUnsetenv(key)
	return
}

func Clearenv() {
	for _, s := range Environ() {
		for j := 0; j < len(s); j++ {
			if s[j] == '=' {
				Unsetenv(s[0:j])
				break
			}
		}
	}
}
