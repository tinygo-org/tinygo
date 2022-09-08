//go:build rp2040
// +build rp2040

// Implementation based on code located here:
// https://github.com/raspberrypi/pico-sdk/blob/master/src/rp2_common/pico_lwip/random.c

package machine

import (
	"device/rp"
)

const numberOfCycles = 32

// GetRNG returns 32 bits of semi-random data based on ring oscillator.
func GetRNG() (uint32, error) {
	var val uint32
	for i := 0; i < 4; i++ {
		val = (val << 8) | uint32(roscRandByte())
	}
	return val, nil
}

var randomByte uint8

func roscRandByte() uint8 {
	var poly uint8
	for i := 0; i < numberOfCycles; i++ {
		if randomByte&0x80 != 0 {
			poly = 0x35
		} else {
			poly = 0
		}
		randomByte = ((randomByte << 1) | uint8(rp.ROSC.GetRANDOMBIT()) ^ poly)
		// TODO: delay a little because the random bit is a little slow
	}
	return randomByte
}
