//go:build rp2350

package runtime

import (
	"device/rp"
	"runtime/interrupt"
)

const (
	// On RP2040 each core has a different IRQ number: SIO_IRQ_PROC0 and SIO_IRQ_PROC1.
	// On RP2350 both cores share the same irq number (SIO_IRQ_PROC) just with a
	// different SIO interrupt output routed to that IRQ input on each core.
	// https://www.raspberrypi.com/documentation/pico-sdk/high_level.html#group_pico_multicore_1ga1413ebfa65114c6f408f4675897ac5ee
	sioIrqFifoProc0 = rp.IRQ_SIO_IRQ_FIFO
	sioIrqFifoProc1 = rp.IRQ_SIO_IRQ_FIFO
)

var sioFifoInterrupt = interrupt.New(sioIrqFifoProc0, handleSIOFifoInterrupt)

// On RP2350 both cores share IRQ_SIO_IRQ_FIFO, but the NVIC enable and
// priority state is per-core. Each core must enable the shared IRQ on
// its own NVIC, so both Core0 and Core1 call Enable()/SetPriority().
func enableSIOFifoInterruptCore0() {
	sioFifoInterrupt.Enable()
	sioFifoInterrupt.SetPriority(0xff)
}

func enableSIOFifoInterruptCore1() {
	sioFifoInterrupt.Enable()
	sioFifoInterrupt.SetPriority(0xff)
}

func handleSIOFifoInterrupt(intr interrupt.Interrupt) {
	switch rp.SIO.FIFO_RD.Get() {
	case 1:
		gcInterruptHandler(currentCPU())
	case 2:
		rp2FlashSafeInterruptHandler(currentCPU())
	}
}
