//go:build baremetal && !(nrf || (stm32 && !(stm32f103 || stm32l0x1)) || (sam && atsamd51) || (sam && atsame5x) || esp32c3 || tkey || (tinygo.riscv32 && virt) || rp2040 || rp2350)

package runtime

func hardwareRand() (n uint64, ok bool) {
	return 0, false // no RNG available
}
