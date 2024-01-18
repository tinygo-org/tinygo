//go:build baremetal && (nrf || (stm32 && !(stm32f103 || stm32l0x1)) || (sam && atsamd51) || (sam && atsame5x) || esp32c3)

// If you update the above build constraint, you'll probably also need to update
// src/crypto/rand/rand_baremetal.go.

package runtime

import "machine"

func hardwareRand() (n uint64, ok bool) {
	n1, err1 := machine.GetRNG()
	n2, err2 := machine.GetRNG()
	n = uint64(n1)<<32 | uint64(n2)
	ok = err1 == nil && err2 == nil
	return
}
