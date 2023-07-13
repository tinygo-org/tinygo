//go:build rp2040

package machine

import (
	"device/rp"
	"runtime/interrupt"
)

// UART on the RP2040.
type UART struct {
	Buffer    *RingBuffer
	Bus       *rp.UART0_Type
	Interrupt interrupt.Interrupt
}

// Configure the UART.
func (uart *UART) Configure(config UARTConfig) error {
	initUART(uart)

	// Default baud rate to 115200.
	if config.BaudRate == 0 {
		config.BaudRate = 115200
	}

	// Use default pins if pins are not set.
	if config.TX == 0 && config.RX == 0 {
		// use default pins
		config.TX = UART_TX_PIN
		config.RX = UART_RX_PIN
	}

	uart.SetBaudRate(config.BaudRate)

	// default to 8-1-N
	uart.SetFormat(8, 1, ParityNone)

	// Enable the UART, both TX and RX
	uart.Bus.UARTCR.SetBits(rp.UART0_UARTCR_UARTEN |
		rp.UART0_UARTCR_RXE |
		rp.UART0_UARTCR_TXE)

	// set GPIO mux to UART for the pins
	if config.TX != NoPin {
		config.TX.Configure(PinConfig{Mode: PinUART})
	}
	if config.RX != NoPin {
		config.RX.Configure(PinConfig{Mode: PinUART})
	}

	// Enable RX IRQ.
	uart.Interrupt.SetPriority(0x80)
	uart.Interrupt.Enable()

	// setup interrupt on receive
	uart.Bus.UARTIMSC.Set(rp.UART0_UARTIMSC_RXIM)

	return nil
}

// SetBaudRate sets the baudrate to be used for the UART.
func (uart *UART) SetBaudRate(br uint32) {
	div := 8 * 125 * MHz / br

	ibrd := div >> 7
	var fbrd uint32

	switch {
	case ibrd == 0:
		ibrd = 1
		fbrd = 0
	case ibrd >= 65535:
		ibrd = 65535
		fbrd = 0
	default:
		fbrd = ((div & 0x7f) + 1) / 2
	}

	// set PL011 baud divisor registers
	uart.Bus.UARTIBRD.Set(ibrd)
	uart.Bus.UARTFBRD.Set(fbrd)

	// PL011 needs a (dummy) line control register write.
	// See https://github.com/raspberrypi/pico-sdk/blob/master/src/rp2_common/hardware_uart/uart.c#L93-L95
	uart.Bus.UARTLCR_H.SetBits(0)
}

// WriteByte writes a byte of data to the UART.
func (uart *UART) writeByte(c byte) error {
	// wait until buffer is not full
	for uart.Bus.UARTFR.HasBits(rp.UART0_UARTFR_TXFF) {
		gosched()
	}

	// write data
	uart.Bus.UARTDR.Set(uint32(c))
	return nil
}

func (uart *UART) flush() {
	for uart.Bus.UARTFR.HasBits(rp.UART0_UARTFR_BUSY) {
		gosched()
	}
}

// SetFormat for number of data bits, stop bits, and parity for the UART.
func (uart *UART) SetFormat(databits, stopbits uint8, parity UARTParity) error {
	var pen, pev uint8
	if parity != ParityNone {
		pen = rp.UART0_UARTLCR_H_PEN
	}
	if parity == ParityEven {
		pev = rp.UART0_UARTLCR_H_EPS
	}
	uart.Bus.UARTLCR_H.SetBits(uint32((databits-5)<<rp.UART0_UARTLCR_H_WLEN_Pos |
		(stopbits-1)<<rp.UART0_UARTLCR_H_STP2_Pos |
		pen | pev))

	return nil
}

func initUART(uart *UART) {
	var resetVal uint32
	switch {
	case uart.Bus == rp.UART0:
		resetVal = rp.RESETS_RESET_UART0
	case uart.Bus == rp.UART1:
		resetVal = rp.RESETS_RESET_UART1
	}

	// reset UART
	rp.RESETS.RESET.SetBits(resetVal)
	rp.RESETS.RESET.ClearBits(resetVal)
	for !rp.RESETS.RESET_DONE.HasBits(resetVal) {
	}
}

// handleInterrupt should be called from the appropriate interrupt handler for
// this UART instance.
func (uart *UART) handleInterrupt(interrupt.Interrupt) {
	for uart.Bus.UARTFR.HasBits(rp.UART0_UARTFR_RXFE) {
	}
	uart.Receive(byte((uart.Bus.UARTDR.Get() & 0xFF)))
}
