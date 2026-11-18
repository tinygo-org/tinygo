//go:build tinygo.riscv && !tkey && !esp32p4

package interrupt

import "device/riscv"

// In returns whether the system is currently in an interrupt.
func In() bool {
	// There is one exception that has the value 0 (instruction address
	// misaligned), but it's not very likely and even if it happens, it's not
	// really something that can be recovered from. Therefore I think it's safe
	// to ignore it. It's handled specially (in handleException).
	return riscv.MCAUSE.Get() != 0
}
