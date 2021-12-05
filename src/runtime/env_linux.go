// +build linux

package runtime

// Update the C environment if cgo is loaded.
// Called from syscall.Setenv.
//go:linkname syscall_setenv_c syscall.setenv_c
func syscall_setenv_c(key string, val string) {
	keydata := cstring(key)
	valdata := cstring(val)
	// ignore any errors
	libc_setenv(&keydata[0], &valdata[0], 1)
	return
}

// cstring converts a Go string to a C string.
// borrowed from syscall
func cstring(s string) []byte {
	data := make([]byte, len(s)+1)
	copy(data, s)
	// final byte should be zero from the initial allocation
	return data
}

// int setenv(const char *name, const char *val, int replace);
//export setenv
func libc_setenv(name *byte, val *byte, replace int32) int32
