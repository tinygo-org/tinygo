//go:build esp32c3 || esp32c6 || esp32h2 || esp32s3

package runtime

import (
	"device/esp"
	"runtime/volatile"
	"unsafe"
)

// interruptMap returns the interrupt matrix register for interrupt source irq.
func interruptMap(irq int) *volatile.Register32 {
	return (*volatile.Register32)(unsafe.Add(unsafe.Pointer(&esp.INTERRUPT_CORE0.CORE_0_INTR_MAP0), irq*4))
}
