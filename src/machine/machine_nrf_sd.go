//go:build nrf && softdevice

package machine

import (
	"device/arm"
	"device/nrf"

	"errors"
)

// avoid a heap allocation in GetRNG.
var (
	bytesAvailable uint8
	buf            [4]uint8

	errNoSoftDeviceSupport = errors.New("rng: softdevice not supported on this device")
	errNotEnoughRandomData = errors.New("rng: not enough random data available")
)

// GetRNG returns 32 bits of non-deterministic random data based on internal thermal noise.
// According to Nordic's documentation, the random output is suitable for cryptographic purposes.
func GetRNG() (ret uint32, err error) {
	// First check whether the SoftDevice is enabled.
	// sd_rand_application_bytes_available_get cannot be called when the SoftDevice is not enabled.
	if !isSoftDeviceEnabled() {
		return getRNG()
	}

	// call into the SoftDevice to get random data bytes available
	switch nrf.Device {
	case "nrf51":
		// sd_rand_application_bytes_available_get: SOC_SVC_BASE_NOT_AVAILABLE + 4
		arm.SVCall1(0x2B+4, &bytesAvailable)
	case "nrf52", "nrf52840", "nrf52833":
		// sd_rand_application_bytes_available_get: SOC_SVC_BASE_NOT_AVAILABLE + 4
		arm.SVCall1(0x2C+4, &bytesAvailable)
	default:
		return 0, errNoSoftDeviceSupport
	}

	if bytesAvailable < 4 {
		return 0, errNotEnoughRandomData
	}

	switch nrf.Device {
	case "nrf51":
		// sd_rand_application_vector_get: SOC_SVC_BASE_NOT_AVAILABLE + 5
		arm.SVCall2(0x2B+5, &buf, 4)
	case "nrf52", "nrf52840", "nrf52833":
		// sd_rand_application_vector_get: SOC_SVC_BASE_NOT_AVAILABLE + 5
		arm.SVCall2(0x2C+5, &buf, 4)
	}

	return uint32(buf[0]) | uint32(buf[1])<<8 | uint32(buf[2])<<16 | uint32(buf[3])<<24, nil
}

// This function is defined in the runtime, but we need it too.
//
//go:linkname isSoftDeviceEnabled runtime.isSoftDeviceEnabled
func isSoftDeviceEnabled() bool
