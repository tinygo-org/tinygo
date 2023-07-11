//go:build darwin || tinygo.wasm

// This implementation of crypto/rand uses the arc4random_buf function
// (available on both MacOS and WASI) to generate random numbers.
//
// Note: arc4random_buf (unlike what the name suggets) does not use the insecure
// RC4 cipher. Instead, it uses a high-quality cipher, varying by the libc
// implementation.

package rand

import "unsafe"

func init() {
	Reader = &reader{}
}

type reader struct {
}

func (r *reader) Read(b []byte) (n int, err error) {
	if len(b) != 0 {
		libc_arc4random_buf(unsafe.Pointer(&b[0]), uint(len(b)))
	}
	return len(b), nil
}

// void arc4random_buf(void *buf, size_t buflen);
//
//export arc4random_buf
func libc_arc4random_buf(buf unsafe.Pointer, buflen uint)
