//go:build rp2040 || nrf || sam

package machine

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
var _ = (func() []byte)(DeviceID)
