//go:build rp2040 || rp2350

// Implementation based on code located here:
// https://github.com/raspberrypi/pico-sdk/blob/master/src/rp2_common/pico_lwip/random.c

package machine

import (
	"device/rp"
)

const numberOfCycles = 32

// GetRNG returns 32 bits of semi-random data based on ring oscillator.
//
// It is safe to call from multiple goroutines.
//
// Unlike some other implementations of GetRNG, these random numbers are not
// cryptographically secure and must not be used for cryptographic operations
// (nonces, etc).
func GetRNG() (uint32, error) {
	core := CurrentCore()
	randomByte := randomBytes[core]
	var val uint32
	for i := 0; i < 4; i++ {
		randomByte = roscRandByte(randomByte)
		val = (val << 8) | uint32(randomByte)
	}
	randomBytes[core] = randomByte
	return val, nil
}

// randomBytes holds the LFSR state for each core.
var randomBytes [2]uint8

func roscRandByte(randomByte uint8) uint8 {
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
