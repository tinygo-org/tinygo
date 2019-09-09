// +build rpi3

package rpi3

// derived from the execellent tutorial by bzt
// https://github.com/bztsrc/raspi3-tutorial/

import "unsafe"
import "runtime/volatile"

func align16bytes(ptr uintptr) uintptr {
	return (ptr + 15) &^ 15
}

// MiniUARTInit sets up the GPIO pins to be MiniUart1 and the settings to be 115200, 8N1
func MiniUARTInit() {
	var r uint32

	/* initialize UART */
	volatile.StoreUint32((*uint32)(AUX_ENABLE), volatile.LoadUint32((*uint32)(AUX_ENABLE))|1) // enable UART1, AUX mini uart
	volatile.StoreUint32((*uint32)(AUX_MU_CNTL), 0)
	volatile.StoreUint32((*uint32)(AUX_MU_LCR), 3) //8 bits?
	volatile.StoreUint32((*uint32)(AUX_MU_MCR), 0)
	volatile.StoreUint32((*uint32)(AUX_MU_IER), 0)
	volatile.StoreUint32((*uint32)(AUX_MU_IIR), 0xc6) //disable interrupts
	volatile.StoreUint32((*uint32)(AUX_MU_BAUD), 270) // 115200 baud
	/* map UART1 to GPIO pins */
	r = volatile.LoadUint32((*uint32)(GPFSEL1))
	r &= ^(uint32(uint32((7 << 12) | (7 << 15)))) // gpio14, gpio15
	r |= (2 << 12) | (2 << 15)                    // alt5
	volatile.StoreUint32((*uint32)(GPFSEL1), r)
	volatile.StoreUint32((*uint32)(GPPUD), 0) // enable pins 14 and 15
	WaitCycles(150)
	volatile.StoreUint32((*uint32)(GPPUDCLK0), (1<<14)|(1<<15))
	WaitCycles(150)
	volatile.StoreUint32((*uint32)(GPPUDCLK0), 0)   // flush GPIO setup
	volatile.StoreUint32((*uint32)(AUX_MU_CNTL), 3) // enable Tx, Rx
}

// UARTSend sends a single  character to UART1
func MiniUARTSend(c byte) {
	/* wait until we can send */
	a := uint32(0)
	//loop until that line goes high
	for volatile.LoadUint32((*uint32)(AUX_MU_LSR))&0x20 == 0 {
		volatile.StoreUint32((*uint32)(&a), volatile.LoadUint32(&a)+1)
	}
	//do{asm volatile("nop");}while(!(*AUX_MU_LSR&0x20));
	/* write the character to the buffer */
	volatile.StoreUint32((*uint32)(AUX_MU_IO), uint32(c))
}

// UARTGetc receives a character (single byte) from the serial port UART1
func MiniUARTGetc() byte {
	var r byte
	for {
		if (volatile.LoadUint32((*uint32)(AUX_MU_LSR)) & 0x01) == 1 {
			break
		}
	}
	r = byte(volatile.LoadUint32((*uint32)(AUX_MU_IO)))
	if r == '\r' {
		return '\n'
	}
	return r
}

// UARTPuts sends a string to UART1 (one character at a time)
func MiniUARTPuts(s string) {
	for _, c := range []byte(s) {
		if c == '\n' {
			MiniUARTSend('\r')
			continue
		}
		MiniUARTSend(c)
	}
}

//this is really a C-style array with 36 elements of type uint32
var mboxData [256]byte

func UART0Init() {
	volatile.StoreUint32((*uint32)(UART0_CR), 0) // turn off UART0

	//32 bit units
	mbox := align16bytes(uintptr(unsafe.Pointer(&mboxData)))

	/* set up clock for consistent divisor values */
	volatile.StoreUint32((*uint32)(unsafe.Pointer(uintptr(mbox)+uintptr(0*4))), 9*4)
	volatile.StoreUint32((*uint32)(unsafe.Pointer(uintptr(mbox)+uintptr(1*4))), MBOX_REQUEST)
	volatile.StoreUint32((*uint32)(unsafe.Pointer(uintptr(mbox)+uintptr(2*4))), MBOX_TAG_SETCLKRATE)
	volatile.StoreUint32((*uint32)(unsafe.Pointer(uintptr(mbox)+uintptr(3*4))), 12)
	volatile.StoreUint32((*uint32)(unsafe.Pointer(uintptr(mbox)+uintptr(4*4))), 8)
	volatile.StoreUint32((*uint32)(unsafe.Pointer(uintptr(mbox)+uintptr(5*4))), 2)
	volatile.StoreUint32((*uint32)(unsafe.Pointer(uintptr(mbox)+uintptr(6*4))), 4000000)
	volatile.StoreUint32((*uint32)(unsafe.Pointer(uintptr(mbox)+uintptr(7*4))), 0)
	volatile.StoreUint32((*uint32)(unsafe.Pointer(uintptr(mbox)+uintptr(8*4))), MBOX_TAG_LAST)
	MboxCall(MBOX_CH_PROP)

	/* map UART0 to GPIO pins */
	var r uint32
	r = volatile.LoadUint32((*uint32)(GPFSEL1))
	r &= ^(uint32(uint32((7 << 12) | (7 << 15)))) // gpio14, gpio15
	r |= (4 << 12) | (4 << 15)                    // alt0
	volatile.StoreUint32((*uint32)(GPFSEL1), r)
	volatile.StoreUint32((*uint32)(GPPUD), 0)

	WaitCycles(150)
	volatile.StoreUint32((*uint32)(GPPUDCLK0), (1<<14)|(1<<15))
	WaitCycles(150)
	volatile.StoreUint32((*uint32)(GPPUDCLK0), 0) // flush GPIO setup

	volatile.StoreUint32((*uint32)(UART0_ICR), 0x7FF) //clear interrupts
	volatile.StoreUint32((*uint32)(UART0_IBRD), 2)    //115200 baud
	volatile.StoreUint32((*uint32)(UART0_FBRD), 0xB)
	volatile.StoreUint32((*uint32)(UART0_LCRH), 3<<5) // 8n1
	volatile.StoreUint32((*uint32)(UART0_CR), 0x301)  // enable Tx, Rx, FIFO
}

/**
 * Make a mailbox call. Returns true if the message has been returned successfully.
 */
func MboxCall(ch byte) bool {
	q := uint32(0)
	mbox := uintptr(unsafe.Pointer(&mboxData)) //64 is bigger than 36 needed, but alloc produces only 4 byte aligned
	//addrMbox := uintptr(unsafe.Pointer(mbox)) & uintptr(f)
	//or on the channel number to he end
	r := (uint32)(mbox | uintptr(uint64(ch)&0xF))
	for volatile.LoadUint32((*uint32)(MBOX_STATUS))&MBOX_FULL != 0 {
		volatile.StoreUint32(&q, volatile.LoadUint32(&q)+1)
	}
	volatile.StoreUint32((*uint32)(MBOX_WRITE), r)
	for {
		for volatile.LoadUint32((*uint32)(MBOX_STATUS))&MBOX_EMPTY != 0 {
			volatile.StoreUint32(&q, volatile.LoadUint32(&q)+1)
		}
		if r == volatile.LoadUint32((*uint32)(MBOX_READ)) {
			resp := volatile.LoadUint32((*uint32)(unsafe.Pointer(uintptr(mbox) + uintptr(1*4))))
			return resp == MBOX_RESPONSE
		}
	}
	return false
}

/**
 * Send a character
 */
func UART0Send(c byte) {
	/* wait until we can send */
	a := uint32(0)
	//loop until that line goes high
	for volatile.LoadUint32((*uint32)(UART0_FR))&0x20 != 0 {
		volatile.StoreUint32(&a, volatile.LoadUint32(&a)+1)
	}
	volatile.StoreUint32((*uint32)(UART0_DR), uint32(c))
}

func UART0DataAvailable() bool {
	return (volatile.LoadUint32((*uint32)(UART0_FR)) & 0x10) == 0
}

func UART0Getc() byte {
	var r byte
	var a = uint32(0)
	for {
		if (volatile.LoadUint32((*uint32)(UART0_FR)) & 0x10) == 0 {
			break
		}
		a = volatile.LoadUint32(&a) + 1
	}
skip:
	r = byte(volatile.LoadUint32((*uint32)(UART0_DR)))
	if r == '\r' {
		goto skip
	}
	return r
}

/**
 * Display a string
 */
func UART0Puts(s string) {

	for _, c := range []byte(s) {
		UART0Send(c)
	}
}

/**
 * Display a binary value in hexadecimal
 */
func UART0Hex(d uint32) {

	for c := 28; c >= 0; c -= 4 {
		// get highest tetrad
		n := (d >> uint32(c)) & 0xF
		// 0-9 => '0'-'9', 10-15 => 'A'-'F'
		if n > 9 {
			n += 0x37
		} else {
			n += 0x30
		}
		UART0Send(byte(n))
	}
	UART0Send('\n')
}

/**
 * Display a 64 bit binary value in hexadecimal
 */
func UART0Hex64(d uint64) {

	for c := 60; c >= 0; c -= 4 {
		// get highest tetrad
		n := (d >> uint32(c)) & 0xF
		// 0-9 => '0'-'9', 10-15 => 'A'-'F'
		if n > 9 {
			n += 0x37
		} else {
			n += 0x30
		}
		UART0Send(byte(n))
	}
	UART0Send('\n')
}

/**
 * Display a binary value in hexadecimal
 */
func MiniUARTHex(d uint32) {
	shiftRight := 28
	for shiftRight >= 0 {
		// get highest tetrad
		n := (d >> uint32(shiftRight)) & 0xF
		// 0-9 => '0'-'9', 10-15 => 'A'-'F'
		if n > 9 {
			n += 0x37
		} else {
			n += 0x30
		}
		MiniUARTSend(byte(n))
		shiftRight -= 4
	}
	MiniUARTSend('\n')
}
