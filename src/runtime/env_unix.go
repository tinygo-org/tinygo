//go:build linux || darwin || wasip1

package runtime

// int setenv(const char *name, const char *val, int replace);
//
//export setenv
func libc_setenv(name *byte, val *byte, replace int32) int32

// int unsetenv(const char *name);
//
//export unsetenv
func libc_unsetenv(name *byte) int32

func setenv(key, val *byte) {
	// ignore any errors
	libc_setenv(key, val, 1)
}

func unsetenv(key *byte) {
	// ignore any errors
	libc_unsetenv(key)
}
