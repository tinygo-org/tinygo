//go:build sam

package machine

import (
	"runtime/volatile"
	"unsafe"
)

var deviceID [16]byte

// DeviceID returns an identifier that is unique within
// a particular chipset.
//
// The identity is one burnt into the MCU itself, or the
// flash chip at time of manufacture.
//
// It's possible that two different vendors may allocate
// the same DeviceID, so callers should take this into
// account if needing to generate a globally unique id.
//
// The length of the hardware ID is vendor-specific, but
// 8 bytes (64 bits) and 16 bytes (128 bits) are common.
func DeviceID() []byte {
	for i := 0; i < len(deviceID); i++ {
		word := (*volatile.Register32)(unsafe.Pointer(deviceIDAddr[i/4])).Get()
		deviceID[i] = byte(word >> ((i % 4) * 8))
	}

	return deviceID[:]
}
