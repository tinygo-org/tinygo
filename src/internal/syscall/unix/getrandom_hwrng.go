//go:build nrf || (stm32 && !(stm32f103 || stm32l0x1 || stm32g0 || stm32u0 || stm32f401)) || (sam && atsamd51) || (sam && atsame5x) || esp32 || esp32c3 || esp32c6 || esp32s3 || tkey || (tinygo.riscv32 && virt)

// If you update the above build constraint, you'll probably also need to update
// src/crypto/rand/rand_baremetal.go and getrandom.go.

package unix

import (
	"syscall"
	_ "unsafe"
)

type GetRandomFlag uintptr

const (
	GRND_NONBLOCK GetRandomFlag = 0x0001
	GRND_RANDOM   GetRandomFlag = 0x0002
)

// GetRandom reads from the hardware RNG. crypto/internal/sysrand uses it for
// the entropy of crypto/mlkem, crypto/ecdh and the other FIPS module packages.
func GetRandom(p []byte, flags GetRandomFlag) (n int, err error) {
	for n < len(p) {
		v, ok := hardwareRand()
		if !ok {
			return n, syscall.EIO
		}
		for i := 0; i < 8 && n < len(p); i++ {
			p[n] = byte(v)
			v >>= 8
			n++
		}
	}
	return n, nil
}

//go:linkname hardwareRand runtime.hardwareRand
func hardwareRand() (n uint64, ok bool)
