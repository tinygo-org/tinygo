//go:build nrf && softdevice
// +build nrf,softdevice

package machine

import (
	"device/arm"
	"device/nrf"
)

//export sd_rand_application_vector_get
func sd_rand_application_vector_get(*[4]uint8, uint8)

// This is a global variable to avoid a heap allocation in GetRNG.
var softdeviceEnabled uint8
var rngData [4]uint8

// GetRNG returns 32 bits of cryptographically secure random data
func GetRNG() (uint32, error) {
	// First check whether the SoftDevice is enabled.
	arm.SVCall1(0x12, &softdeviceEnabled) // sd_softdevice_is_enabled

	if softdeviceEnabled != 0 {
		// Now pick the appropriate SVCall number. Hopefully they won't change
		// in the future with a different SoftDevice version.
		if nrf.Device == "nrf51" {
			// sd_rand_application_vector_get: SOC_SVC_BASE_NOT_AVAILABLE + 15
			arm.SVCall2(0x2B+15, &rngData, 4)
		} else if nrf.Device == "nrf52" || nrf.Device == "nrf52840" || nrf.Device == "nrf52833" {
			// sd_rand_application_vector_get: SOC_SVC_BASE_NOT_AVAILABLE + 5
			arm.SVCall2(0x2C+5, &rngData, 4)
		} else {
			sd_rand_application_vector_get(&rngData, 4)
		}

		result := uint32(rngData[0])<<24 | uint32(rngData[1])<<16 | uint32(rngData[2])<<8 | uint32(rngData[3])
		return result, nil
	}

	// SoftDevice is disabled so we use normal call.
	return getRNG()
}
