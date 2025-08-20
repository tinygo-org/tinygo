//go:build rp2040

package runtime

import (
	"device/rp"
)

const (
	sioIrqFifoProc0 = rp.IRQ_SIO_IRQ_PROC0
	sioIrqFifoProc1 = rp.IRQ_SIO_IRQ_PROC1
)
