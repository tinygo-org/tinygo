//go:build nrf || (stm32 && !(stm32f103 || stm32l0x1)) || (sam && atsamd51) || (sam && atsame5x) || esp32c3

// If you update the above build constraint, you'll probably also need to update
// src/runtime/rand_hwrng.go.

package rand

import (
	"machine"
)

func init() {
	Reader = &reader{}
}

type reader struct {
}

func (r *reader) Read(b []byte) (n int, err error) {
	if len(b) == 0 {
		return
	}

	var randomByte uint32
	for i := range b {
		if i%4 == 0 {
			randomByte, err = machine.GetRNG()
			if err != nil {
				return n, err
			}
		} else {
			randomByte >>= 8
		}
		b[i] = byte(randomByte)
	}

	return len(b), nil
}
