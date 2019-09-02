// +build rpi3

package runtime

// derived from the execellent tutorial by bzt
// https://github.com/bztsrc/raspi3-tutorial/

import "unsafe"
import "runtime/volatile"
import dev "device/rpi3"

// MiniUARTInit sets up the GPIO pins to be MiniUart1 and the settings to be 115200, 8N1
func MiniUARTInit() {
	var r uint32

	/* initialize UART */
	put32(dev.AUX_ENABLE, get32(dev.AUX_ENABLE)|1) // enable UART1, AUX mini uart
	put32(dev.AUX_MU_CNTL, 0)
	put32(dev.AUX_MU_LCR, 3) //8 bits?
	put32(dev.AUX_MU_MCR, 0)
	put32(dev.AUX_MU_IER, 0)
	put32(dev.AUX_MU_IIR, 0xc6) //disable interrupts
	put32(dev.AUX_MU_BAUD, 270) // 115200 baud
	/* map UART1 to GPIO pins */
	r = get32(dev.GPFSEL1)
	r &= ^(uint32(uint32((7 << 12) | (7 << 15)))) // gpio14, gpio15
	r |= (2 << 12) | (2 << 15)                    // alt5
	put32(dev.GPFSEL1, r)
	put32(dev.GPPUD, 0) // enable pins 14 and 15
	a := 0
	for r := 0; r <= 150; r++ {
		a++
	}
	put32(dev.GPPUDCLK0, (1<<14)|(1<<15))
	for r := 0; r <= 150; r++ {
		a++
	}
	put32(dev.GPPUDCLK0, 0)   // flush GPIO setup
	put32(dev.AUX_MU_CNTL, 3) // enable Tx, Rx
}

func get32(ptr unsafe.Pointer) uint32 {
	return volatile.LoadUint32((*uint32)(ptr))
}
func put32(ptr unsafe.Pointer, value uint32) {
	volatile.StoreUint32((*uint32)(ptr), value)
}

// UARTSend sends a single  character to UART1
func MiniUARTSend(c byte) {
	/* wait until we can send */
	a := uint32(0)
	//loop until that line goes high
	for volatile.LoadUint32((*uint32)(dev.AUX_MU_LSR))&0x20 == 0 {
		volatile.StoreUint32((*uint32)(&a), volatile.LoadUint32(&a)+1)
	}
	//do{asm volatile("nop");}while(!(*AUX_MU_LSR&0x20));
	/* write the character to the buffer */
	volatile.StoreUint32((*uint32)(dev.AUX_MU_IO), uint32(c))
}

// UARTGetc receives a character (single byte) from the serial port UART1
func MiniUARTGetc() byte {
	var r byte
	for {
		if (get32(dev.AUX_MU_LSR) & 0x01) == 1 {
			break
		}
	}
	r = byte(get32(dev.AUX_MU_IO))
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
	volatile.StoreUint32((*uint32)(dev.UART0_CR), 0) // turn off UART0

	//32 bit units
	mbox := align(uintptr(unsafe.Pointer(&mboxData)))

	/* set up clock for consistent divisor values */
	volatile.StoreUint32((*uint32)(unsafe.Pointer(uintptr(mbox)+uintptr(0*4))), 9*4)
	volatile.StoreUint32((*uint32)(unsafe.Pointer(uintptr(mbox)+uintptr(1*4))), dev.MBOX_REQUEST)
	volatile.StoreUint32((*uint32)(unsafe.Pointer(uintptr(mbox)+uintptr(2*4))), dev.MBOX_TAG_SETCLKRATE)
	volatile.StoreUint32((*uint32)(unsafe.Pointer(uintptr(mbox)+uintptr(3*4))), 12)
	volatile.StoreUint32((*uint32)(unsafe.Pointer(uintptr(mbox)+uintptr(4*4))), 8)
	volatile.StoreUint32((*uint32)(unsafe.Pointer(uintptr(mbox)+uintptr(5*4))), 2)
	volatile.StoreUint32((*uint32)(unsafe.Pointer(uintptr(mbox)+uintptr(6*4))), 4000000)
	volatile.StoreUint32((*uint32)(unsafe.Pointer(uintptr(mbox)+uintptr(7*4))), 0)
	volatile.StoreUint32((*uint32)(unsafe.Pointer(uintptr(mbox)+uintptr(8*4))), dev.MBOX_TAG_LAST)
	MboxCall(dev.MBOX_CH_PROP)

	/* map UART0 to GPIO pins */
	var r uint32
	r = volatile.LoadUint32((*uint32)(dev.GPFSEL1))
	r &= ^(uint32(uint32((7 << 12) | (7 << 15)))) // gpio14, gpio15
	r |= (4 << 12) | (4 << 15)                    // alt0
	volatile.StoreUint32((*uint32)(dev.GPFSEL1), r)
	volatile.StoreUint32((*uint32)(dev.GPPUD), 0)

	a := 0
	for r := 0; r <= 150; r++ {
		a++
	}
	volatile.StoreUint32((*uint32)(dev.GPPUDCLK0), (1<<14)|(1<<15))
	for r := 0; r <= 150; r++ {
		a++
	}
	volatile.StoreUint32((*uint32)(dev.GPPUDCLK0), 0) // flush GPIO setup

	volatile.StoreUint32((*uint32)(dev.UART0_ICR), 0x7FF) //clear interrupts
	volatile.StoreUint32((*uint32)(dev.UART0_IBRD), 2)    //115200 baud
	volatile.StoreUint32((*uint32)(dev.UART0_FBRD), 0xB)
	volatile.StoreUint32((*uint32)(dev.UART0_LCRH), 3<<5) // 8n1
	volatile.StoreUint32((*uint32)(dev.UART0_CR), 0x301)  // enable Tx, Rx, FIFO
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
	UART0Hex(r)
	for volatile.LoadUint32((*uint32)(dev.MBOX_STATUS))&dev.MBOX_FULL != 0 {
		volatile.StoreUint32(&q, volatile.LoadUint32(&q)+1)
	}
	volatile.StoreUint32((*uint32)(dev.MBOX_WRITE), r)
	for {
		for volatile.LoadUint32((*uint32)(dev.MBOX_STATUS))&dev.MBOX_EMPTY != 0 {
			volatile.StoreUint32(&q, volatile.LoadUint32(&q)+1)
		}
		if r == volatile.LoadUint32((*uint32)(dev.MBOX_READ)) {
			resp := volatile.LoadUint32((*uint32)(unsafe.Pointer(uintptr(mbox) + uintptr(1*4))))
			return resp == dev.MBOX_RESPONSE
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
	for volatile.LoadUint32((*uint32)(dev.UART0_FR))&0x20 != 0 {
		volatile.StoreUint32(&a, volatile.LoadUint32(&a)+1)
	}
	volatile.StoreUint32((*uint32)(dev.UART0_DR), uint32(c))
}

/**
 * Receive a character
 */
func UART0Getc() byte {
	var r byte
	var a = uint32(0)
	for {
		if (volatile.LoadUint32((*uint32)(dev.UART0_FR)) & 0x10) != 0 {
			break
		}
		a = volatile.LoadUint32(&a) + 1
	}
	r = byte(volatile.LoadUint32((*uint32)(dev.UART0_DR)))
	if r == '\r' {
		return '\n'
	}
	return r
}

/**
 * Display a string
 */
func UART0Puts(s string) {

	for _, c := range []byte(s) {
		if c == '\n' {
			UART0Send('\r')
			continue
		}
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
