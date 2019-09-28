// +build rpi3

package machine

import rpi "device/arm64"
import "runtime/volatile"

// PinMode and Pin are not implemented yet. Everything on RPI is basically GPIO and
// under software control.
type PinMode uint8

// Set is not implemented.
func (p Pin) Set(_ bool) {
}

type UART struct {
	Buffer *RingBuffer
}

var (
	MiniUART = UART{Buffer: NewRingBuffer()}
)

//
// Configure ignores the provided params and always does 8,N,1 on pins 14 and 15.
// The pins and UART mapped to them are under software control, so it's not really
// possible to have a static map of the pins.
//
func (uart UART) Configure(_ UARTConfig) {
	var r uint32

	/* initialize UART */
	volatile.StoreUint32((*uint32)(rpi.AUX_ENABLE), volatile.LoadUint32((*uint32)(rpi.AUX_ENABLE))|1) //enable mini UART
	volatile.StoreUint32((*uint32)(rpi.AUX_MU_CNTL), 0)
	volatile.StoreUint32((*uint32)(rpi.AUX_MU_LCR), 3) //8 bits?
	volatile.StoreUint32((*uint32)(rpi.AUX_MU_MCR), 0)
	volatile.StoreUint32((*uint32)(rpi.AUX_MU_IER), 0)
	volatile.StoreUint32((*uint32)(rpi.AUX_MU_IIR), 0xc6) //disable interrupts
	volatile.StoreUint32((*uint32)(rpi.AUX_MU_BAUD), 270) // 115200 baud
	/* map UART1 to GPIO pins */
	r = volatile.LoadUint32((*uint32)(rpi.GPFSEL1))
	r &= ^(uint32(uint32((7 << 12) | (7 << 15)))) // gpio14, gpio15
	r |= (2 << 12) | (2 << 15)                    // alt5
	volatile.StoreUint32((*uint32)(rpi.GPFSEL1), r)
	volatile.StoreUint32((*uint32)(rpi.GPPUD), 0) // enable pins 14 and 15

	// we would like to call runtime.WaitCycles here but that would form an
	// import loop
	n := 150
	for n > 0 {
		rpi.Asm("nop")
		n--
	}
	volatile.StoreUint32((*uint32)(rpi.GPPUDCLK0), 0)   // flush GPIO setup
	volatile.StoreUint32((*uint32)(rpi.AUX_MU_CNTL), 3) // enable Tx, Rx
}

func (uart UART) WriteByte(v byte) {
	/* wait until we can send */
	a := uint32(0)
	//loop until that line goes high
	for volatile.LoadUint32((*uint32)(rpi.AUX_MU_LSR))&0x20 == 0 {
		volatile.StoreUint32((*uint32)(&a), volatile.LoadUint32(&a)+1)
	}
	//do{asm volatile("nop");}while(!(*rpi.AUX_MU_LSR&0x20));
	/* write the character to the buffer */
	volatile.StoreUint32((*uint32)(rpi.AUX_MU_IO), uint32(v))
}
