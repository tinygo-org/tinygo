//go:build stm32f4 || stm32f7 || stm32l4 || stm32wlx
// +build stm32f4 stm32f7 stm32l4 stm32wlx

package machine

import "device/stm32"

var rngInitDone = false

const RNG_MAX_READ_RETRIES = 1000

// GetRNG returns 32 bits of cryptographically secure random data
func GetRNG() (uint32, error) {
	if !rngInitDone {
		initRNG()
		rngInitDone = true
	}

	cnt := RNG_MAX_READ_RETRIES
	for !stm32.RNG.SR.HasBits(stm32.RNG_SR_DRDY) {
		cnt--
		if cnt == 0 {
			return 0, ErrTimeoutRNG
		}
	}

	ret := stm32.RNG.DR.Get()
	return ret, nil
}
