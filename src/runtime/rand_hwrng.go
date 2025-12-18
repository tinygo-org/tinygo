//go:build baremetal && (nrf || (stm32 && !(stm32f103 || stm32l0x1)) || (sam && atsamd51) || (sam && atsame5x) || esp32c3 || tkey || (tinygo.riscv32 && virt) || rp2040 || rp2350)

// If you update the above build constraint, you'll probably also need to update
// src/crypto/rand/rand_baremetal.go.
//
// The rp2040 and rp2350 implementations are not included in src/crypto/rand/rand_baremetal.go
// due to not being sufficiently random for the Go crypto libs.
// However since the randomness here does not provide those same guarantees,
// they are included in the list for hardwareRand() implementations.

package runtime

import "machine"

func hardwareRand() (n uint64, ok bool) {
	n1, err1 := machine.GetRNG()
	n2, err2 := machine.GetRNG()
	n = uint64(n1)<<32 | uint64(n2)
	ok = err1 == nil && err2 == nil
	return
}
