package rand

import (
	"errors"
	"unsafe"
)

func init() {
	Reader = &reader{}
}

type reader struct {
}

var errRandom = errors.New("failed to obtain random data from rand_s")

func (r *reader) Read(b []byte) (n int, err error) {
	if len(b) == 0 {
		return
	}

	// Use the old RtlGenRandom, introduced in Windows XP.
	// Even though the documentation says it is deprecated, it is widely used
	// and probably won't go away anytime soon.
	// See for example: https://github.com/golang/go/issues/33542
	// For Windows 7 and newer, we might switch to ProcessPrng in the future
	// (which is a documented function and might be a tiny bit faster).
	ok := libc_RtlGenRandom(unsafe.Pointer(&b[0]), len(b))
	if !ok {
		return 0, errRandom
	}
	return len(b), nil
}

// This function is part of advapi32.dll, and is called SystemFunction036 for
// some reason. It's available on Windows XP and newer.
// See: https://learn.microsoft.com/en-us/windows/win32/api/ntsecapi/nf-ntsecapi-rtlgenrandom
//
//export SystemFunction036
func libc_RtlGenRandom(buf unsafe.Pointer, len int) bool
