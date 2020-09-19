// +build wasm

package wasi

import "unsafe"

// Implements __wasi_ciovec_t and __wasi_iovec_t.
type IOVec struct {
	Buf    unsafe.Pointer
	BufLen uint
}

//go:wasm-module wasi_unstable
//export fd_write
func Fd_write(id uint32, iovs *IOVec, iovs_len uint, nwritten *uint) (errno uint)
