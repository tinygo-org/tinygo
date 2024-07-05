// Code generated by wit-bindgen-go. DO NOT EDIT.

//go:build !wasip1

package ipnamelookup

import (
	"github.com/ydnar/wasm-tools-go/cm"
	"internal/wasi/sockets/v0.2.0/network"
	"unsafe"
)

// OptionIPAddressShape is used for storage in variant or result types.
type OptionIPAddressShape struct {
	shape [unsafe.Sizeof(cm.Option[network.IPAddress]{})]byte
}
