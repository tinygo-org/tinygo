// Code generated by wit-bindgen-go. DO NOT EDIT.

package ipnamelookup

import (
	"internal/cm"
	"internal/wasi/sockets/v0.2.0/network"
	"unsafe"
)

// OptionIPAddressShape is used for storage in variant or result types.
type OptionIPAddressShape struct {
	shape [unsafe.Sizeof(cm.Option[network.IPAddress]{})]byte
}
