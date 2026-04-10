//go:build !tinygo.wasm

package unix

import "syscall"

type GetRandomFlag uintptr

const (
	GRND_NONBLOCK GetRandomFlag = 0x0001
	GRND_RANDOM   GetRandomFlag = 0x0002
)

func GetRandom(p []byte, flags GetRandomFlag) (n int, err error) {
	// Not supported on most TinyGo targets.
	// On real Linux the sysrand package will fall back to /dev/urandom.
	return 0, syscall.ENOSYS
}
