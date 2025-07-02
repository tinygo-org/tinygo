package unix

import "syscall"

func Eaccess(path string, mode uint32) error {
	// We don't support this syscall on baremetal or wasm.
	// Callers are generally able to deal with this since unix.Eaccess also
	// isn't available on Android.
	return syscall.ENOSYS
}
