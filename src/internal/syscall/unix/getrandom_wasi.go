//go:build tinygo.wasm

package unix

import "unsafe"

type GetRandomFlag uintptr

const (
	GRND_NONBLOCK GetRandomFlag = 0x0001
	GRND_RANDOM   GetRandomFlag = 0x0002
)

func GetRandom(p []byte, flags GetRandomFlag) (n int, err error) {
	if len(p) == 0 {
		return 0, nil
	}
	libc_arc4random_buf(unsafe.Pointer(unsafe.SliceData(p)), uint(len(p)))
	return len(p), nil
}

// void arc4random_buf(void *buf, size_t buflen);
//
//export arc4random_buf
func libc_arc4random_buf(buf unsafe.Pointer, buflen uint)
