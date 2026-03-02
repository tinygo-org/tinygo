//go:build esp32s3 || esp32c3

package machine

import (
	"runtime/volatile"
)

// SPIConfig is used to store config info for SPI.
type SPIConfig struct {
	Frequency uint32
	SCK       Pin   // Serial Clock
	SDO       Pin   // Serial Data Out (MOSI)
	SDI       Pin   // Serial Data In  (MISO)
	CS        Pin   // Chip Select (optional)
	LSBFirst  bool  // MSB is default
	Mode      uint8 // Mode0 is default
}

// freqToClockDiv computes the SPI bus clock divider register value.
// SPI peripherals on ESP32-C3 and ESP32-S3 are clocked from the APB bus
// (pplClockFreq, 80 MHz on both chips).
func freqToClockDiv(hz uint32) uint32 {
	if hz >= pplClockFreq { // maximum frequency
		return 1 << 31
	}
	if hz < (pplClockFreq / (16 * 64)) { // minimum frequency
		return 15<<18 | 63<<12 | 31<<6 | 63 // pre=15, n=63
	}

	// Iterate all 16 prescaler options looking for an exact match
	// or the smallest error.
	var bestPre, bestN, bestErr uint32
	bestN = 1
	bestErr = 0xffffffff
	q := uint32(float32(pplClockFreq)/float32(hz) + float32(0.5))
	for p := uint32(0); p < 16; p++ {
		n := q/(p+1) - 1
		if n < 1 { // prescaler became too large, stop enum
			break
		}
		if n > 63 { // prescaler too small, skip to next
			continue
		}

		freq := pplClockFreq / ((p + 1) * (n + 1))
		if freq == hz { // exact match
			return p<<18 | n<<12 | (n/2)<<6 | n
		}

		var err uint32
		if freq < hz {
			err = hz - freq
		} else {
			err = freq - hz
		}
		if err < bestErr {
			bestErr = err
			bestPre = p
			bestN = n
		}
	}

	return bestPre<<18 | bestN<<12 | (bestN/2)<<6 | bestN
}

// spiTxFillBuffer writes data from w into the 16-word (64-byte) SPI
// hardware transfer buffer. Unused words are zeroed so that no stale
// data from a previous transfer is sent when w is shorter than 64 bytes.
func spiTxFillBuffer(buf *[16]volatile.Register32, w []byte) {
	if len(w) >= 64 {
		// We can fill the entire 64-byte transfer buffer with data.
		// This loop is slightly faster than the loop below.
		for i := 0; i < 16; i++ {
			word := uint32(w[i*4]) | uint32(w[i*4+1])<<8 | uint32(w[i*4+2])<<16 | uint32(w[i*4+3])<<24
			buf[i].Set(word)
		}
	} else {
		// We can't fill the entire transfer buffer, so we need to be a bit
		// more careful.
		// Note that parts of the transfer buffer that aren't used still
		// need to be set to zero, otherwise we might be transferring
		// garbage from a previous transmission if w is smaller than r.
		for i := 0; i < 16; i++ {
			var word uint32
			if i*4+3 < len(w) {
				word |= uint32(w[i*4+3]) << 24
			}
			if i*4+2 < len(w) {
				word |= uint32(w[i*4+2]) << 16
			}
			if i*4+1 < len(w) {
				word |= uint32(w[i*4+1]) << 8
			}
			if i*4+0 < len(w) {
				word |= uint32(w[i*4+0]) << 0
			}
			buf[i].Set(word)
		}
	}
}
