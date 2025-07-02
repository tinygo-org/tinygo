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
			src := *(*[]byte)(unsafe.Pointer(&sliceHeader{buf: raw, len: size, cap: size}))
			return string(src), true
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

//go:extern environ
var libc_environ *unsafe.Pointer
