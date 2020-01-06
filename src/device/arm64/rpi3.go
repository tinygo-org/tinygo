// +build rpi3

package arm64

import "unsafe"

//
// NOTE: This value is ONLY for the RPI2 and RPI3.  The RPI4 uses 0x40000000.
//
var MMIO_BASE = unsafe.Pointer(uintptr(0x3F000000))
