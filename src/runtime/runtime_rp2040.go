//go:build rp2040

package runtime

import (
	"device/rp"
	"runtime/interrupt"
)

const (
	sioIrqFifoProc0 = rp.IRQ_SIO_IRQ_PROC0
	sioIrqFifoProc1 = rp.IRQ_SIO_IRQ_PROC1
)

// On RP2040, each core has its own SIO FIFO IRQ. Core0 enables
// IRQ_SIO_IRQ_PROC0 and Core1 enables IRQ_SIO_IRQ_PROC1, so each handler can
// use a fixed core ID.
func enableSIOFifoInterruptCore0() {
	intr := interrupt.New(sioIrqFifoProc0, handleSIOFifoInterruptCore0)
	intr.Enable()
	intr.SetPriority(0xff)
}

func enableSIOFifoInterruptCore1() {
	intr := interrupt.New(sioIrqFifoProc1, handleSIOFifoInterruptCore1)
	intr.Enable()
	intr.SetPriority(0xff)
}

func handleSIOFifoInterruptCore0(intr interrupt.Interrupt) {
	switch rp.SIO.FIFO_RD.Get() {
	case 1:
		gcInterruptHandler(0)
	case 2:
		rp2FlashSafeInterruptHandler(0)
	}
}

func handleSIOFifoInterruptCore1(intr interrupt.Interrupt) {
	switch rp.SIO.FIFO_RD.Get() {
	case 1:
		gcInterruptHandler(1)
	case 2:
		rp2FlashSafeInterruptHandler(1)
	}
}
