//go:build nrf52840
// +build nrf52840

package machine

import (
	"device/nrf"
)

// Implementation based on Nordic Semiconductor's nRF52840 documentation version 1.7 found here:
// https://infocenter.nordicsemi.com/pdf/nRF52840_PS_v1.7.pdf

// SetRNGBiasCorrection configures the RNG peripheral's bias correction mechanism. Note that when
// bias correction is enabled, the peripheral is slower to produce random values.
func SetRNGBiasCorrection(enabled bool) {
	var val uint32
	if enabled {
		val = nrf.RNG_CONFIG_DERCEN_Enabled
	}
	nrf.RNG.SetCONFIG_DERCEN(val)
}

// RNGBiasCorrectionEnabled determines whether the RNG peripheral's bias correction mechanism is
// enabled or not.
func RNGBiasCorrectionEnabled() bool {
	return nrf.RNG.GetCONFIG_DERCEN() == nrf.RNG_CONFIG_DERCEN_Enabled
}

// StartRNG starts the RNG peripheral core. This is automatically called by GetRNG, but can be
// manually called for interacting with the RNG peripheral directly.
func StartRNG() {
	nrf.RNG.SetTASKS_START(nrf.RNG_TASKS_START_TASKS_START_Trigger)
}

// StopRNG stops the RNG peripheral core. This is not called automatically. It may make sense to
// manually disable RNG peripheral for power conservation.
func StopRNG() {
	nrf.RNG.SetTASKS_STOP(nrf.RNG_TASKS_STOP_TASKS_STOP_Trigger)
}

// GetRNG returns 32 bits of non-deterministic random data based on internal thermal noise.
// According to Nordic's documentation, the random output is suitable for cryptographic purposes.
func GetRNG() (ret uint32, err error) {
	// There's no apparent way to check the status of the RNG peripheral's task, so simply start it
	// to avoid deadlocking while waiting for output.
	StartRNG()

	// The RNG returns one byte at a time, so stack up four bytes into a single uint32 for return.
	for i := 0; i < 4; i++ {
		// Wait for data to be ready.
		for nrf.RNG.GetEVENTS_VALRDY() == nrf.RNG_EVENTS_VALRDY_EVENTS_VALRDY_NotGenerated {
		}
		// Append random byte to output.
		ret = (ret << 8) ^ nrf.RNG.GetVALUE()
		// Unset the EVENTS_VALRDY register to avoid reading the same random output twice.
		nrf.RNG.SetEVENTS_VALRDY(nrf.RNG_EVENTS_VALRDY_EVENTS_VALRDY_NotGenerated)
	}

	return ret, nil
}
