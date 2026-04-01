//go:build tinygo.riscv && esp32c3_qemu_target

package interrupt

// Enable is a no-op stub for the QEMU esp32c3 test target.
// Hardware interrupts are not needed when running tests under qemu-system-riscv32.
func (i Interrupt) Enable() error { return nil }
