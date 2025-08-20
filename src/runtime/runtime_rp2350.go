//go:build rp2350

package runtime

import (
	"device/rp"
)

const (
	// On RP2040 each core has a different IRQ number: SIO_IRQ_PROC0 and SIO_IRQ_PROC1.
	// On RP2350 both cores share the same irq number (SIO_IRQ_PROC) just with a
	// different SIO interrupt output routed to that IRQ input on each core.
	// https://www.raspberrypi.com/documentation/pico-sdk/high_level.html#group_pico_multicore_1ga1413ebfa65114c6f408f4675897ac5ee
	sioIrqFifoProc0 = rp.IRQ_SIO_IRQ_FIFO
	sioIrqFifoProc1 = rp.IRQ_SIO_IRQ_FIFO
)
