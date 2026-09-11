//go:build espidf

package rand

import "unsafe"

func init() {
	Reader = &reader{}
}

type reader struct{}

//export esp_fill_random
func esp_fill_random(buf unsafe.Pointer, len uintptr)

func (r *reader) Read(b []byte) (n int, err error) {
	if len(b) == 0 {
		return
	}
	esp_fill_random(unsafe.Pointer(&b[0]), uintptr(len(b)))
	return len(b), nil
}
