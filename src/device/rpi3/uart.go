// +build rpi3

package rpi3

// derived from the execellent tutorial by bzt
// https://github.com/bztsrc/raspi3-tutorial/

import "unsafe"

// UARTInit sets up the GPIO pins to be UART1 and the settings to be 115200, 8N1
func UARTInit() {
	var r uint32

	/* initialize UART */
	put32(AUX_ENABLE, get32(AUX_ENABLE)|1) // enable UART1, AUX mini uart
	put32(AUX_MU_CNTL, 0)
	put32(AUX_MU_LCR, 3) //8 bits?
	put32(AUX_MU_MCR, 0)
	put32(AUX_MU_IER, 0)
	put32(AUX_MU_IIR, 0xc6) //disable interrupts
	put32(AUX_MU_BAUD, 270) // 115200 baud
	/* map UART1 to GPIO pins */
	r = get32(GPFSEL1)
	r &= ^(uint32(uint32((7 << 12) | (7 << 15)))) // gpio14, gpio15
	r |= (2 << 12) | (2 << 15)                    // alt5
	put32(GPFSEL1, r)
	put32(GPPUD, 0) // enable pins 14 and 15
	a := 0
	for r := 0; r <= 150; r++ {
		a++
	}
	put32(GPPUDCLK0, (1<<14)|(1<<15))
	for r := 0; r <= 150; r++ {
		a++
	}
	put32(GPPUDCLK0, 0)   // flush GPIO setup
	put32(AUX_MU_CNTL, 3) // enable Tx, Rx
}

func get32(ptr unsafe.Pointer) uint32 {
	cast := (*uint32)(ptr)
	return *cast
}
func put32(ptr unsafe.Pointer, value uint32) {
	cast := (*uint32)(ptr)
	*cast = value
}

// UARTSend sends a single  character to UART1
func UARTSend(c byte) {
	/* wait until we can send */
	a := 0
	//loop until that line goes high
	for get32(AUX_MU_LSR)&0x20 == 0 {
		a++
	}
	//do{asm volatile("nop");}while(!(*AUX_MU_LSR&0x20));
	/* write the character to the buffer */
	put32(AUX_MU_IO, uint32(c))
}

// UARTGetc receives a character (single byte) from the serial port UART1
func UARTGetc() byte {
	var r byte
	for {
		if (get32(AUX_MU_LSR) & 0x01) == 1 {
			break
		}
	}
	r = byte(get32(AUX_MU_IO))
	if r == '\r' {
		return '\n'
	}
	return r
}

// UARTPuts sends a string to UART1 (one character at a time)
func UARTPuts(s string) {
	for _, c := range []byte(s) {
		if c == '\n' {
			UARTSend('\r')
			continue
		}
		UARTSend(c)
	}
}
