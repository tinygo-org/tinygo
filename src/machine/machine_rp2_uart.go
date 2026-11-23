//go:build (rp2040 || rp2350) && machine.uart

package machine

import (
	"device/rp"
	"errors"
	"runtime/interrupt"
)

var (
	UART0  = &_UART0
	_UART0 = UART{
		Buffer: NewRingBuffer(),
		Bus:    rp.UART0,
	}

	UART1  = &_UART1
	_UART1 = UART{
		Buffer: NewRingBuffer(),
		Bus:    rp.UART1,
	}

	errSamePins         = errors.New("UART: invalid pin combination")
	errWrongBitSize     = errors.New("UART: invalid data size")
	errWrongStopBitSize = errors.New("UART: invalid bit size")
)

func init() {
	UART0.Interrupt = interrupt.New(rp.IRQ_UART0_IRQ, _UART0.handleInterrupt)
	UART1.Interrupt = interrupt.New(rp.IRQ_UART1_IRQ, _UART1.handleInterrupt)
}

// UART on the RP2040.
type UART struct {
	Buffer    *RingBuffer
	Bus       *rp.UART0_Type
	Interrupt interrupt.Interrupt
}

// Configure the UART.
func (uart *UART) Configure(config UARTConfig) error {
	// Use default pins if pins are not set.
	if config.TX == 0 && config.RX == 0 {
		// use default pins
		config.TX = UART_TX_PIN
		config.RX = UART_RX_PIN
	}
	if config.TX == config.RX {
		return errSamePins
	}

	initUART(uart)

	uart.SetBaudRate(config.BaudRate)

	// default to 8-1-N
	uart.SetFormat(8, 1, ParityNone)

	// Enable the UART, both TX and RX
	settings := uint32(rp.UART0_UARTCR_UARTEN |
		rp.UART0_UARTCR_RXE |
		rp.UART0_UARTCR_TXE)
	const bits = rp.UART0_UARTCR_UARTEN | rp.UART0_UARTCR_TXE
	if config.RTS != 0 {
		settings |= rp.UART0_UARTCR_RTSEN
	}
	if config.CTS != 0 {
		settings |= rp.UART0_UARTCR_CTSEN
	}

	uart.Bus.UARTCR.SetBits(settings)

	// set GPIO mux to UART for the pins
	if config.TX != NoPin {
		config.TX.Configure(PinConfig{Mode: PinUART})
		if config.InvertTX {
			config.TX.setOutOver(rp.IO_BANK0_GPIO0_CTRL_OUTOVER_INVERT)
		} else {
			config.TX.setOutOver(rp.IO_BANK0_GPIO0_CTRL_OUTOVER_NORMAL)
		}
	}
	if config.RX != NoPin {
		config.RX.Configure(PinConfig{Mode: PinUART})
		if config.InvertRX {
			config.RX.setInOver(rp.IO_BANK0_GPIO0_CTRL_INOVER_INVERT)
		} else {
			config.RX.setInOver(rp.IO_BANK0_GPIO0_CTRL_INOVER_NORMAL)
		}
	}
	if config.RTS != 0 {
		config.RTS.Configure(PinConfig{Mode: PinOutput})
	}
	if config.CTS != 0 {
		config.CTS.Configure(PinConfig{Mode: PinInput})
	}

	// Enable RX IRQ.
	uart.Interrupt.SetPriority(0x80)
	uart.Interrupt.Enable()

	// Setup interrupt on receive.
	uart.Bus.UARTIMSC.Set(rp.UART0_UARTIMSC_RXIM)

	return nil
}

// Close the UART and disable its interrupt/power use.
func (uart *UART) Close() error {
	uart.Interrupt.Disable()

	// Disable UART.
	uart.Bus.UARTCR.ClearBits(rp.UART0_UARTCR_UARTEN)

	var resetVal uint32
	switch {
	case uart.Bus == rp.UART0:
		resetVal = rp.RESETS_RESET_UART0
	case uart.Bus == rp.UART1:
		resetVal = rp.RESETS_RESET_UART1
	}

	// reset UART
	resetBlock(resetVal)

	return nil
}

// SetBaudRate sets the baudrate to be used for the UART.
func (uart *UART) SetBaudRate(br uint32) {
	if br == 0 {
		br = 115200
	}
	div := 8 * CPUFrequency() / br

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
		if !interrupt.In() {
			gosched()
		}
	}

	// write data
	uart.Bus.UARTDR.Set(uint32(c))
	return nil
}

func (uart *UART) flush() {
	for uart.Bus.UARTFR.HasBits(rp.UART0_UARTFR_BUSY) {
		if !interrupt.In() {
			gosched()
		}
	}
}

// SetFormat for number of data bits, stop bits, and parity for the UART.
func (uart *UART) SetFormat(databits, stopbits uint8, parity UARTParity) error {
	if databits < 5 || databits > 8 {
		return errWrongBitSize
	}
	if stopbits < 1 || stopbits > 2 {
		return errWrongStopBitSize
	}
	lcr := uint32(databits-5) << rp.UART0_UARTLCR_H_WLEN_Pos
	if stopbits == 2 {
		lcr |= rp.UART0_UARTLCR_H_STP2
	}
	if parity != ParityNone {
		lcr |= rp.UART0_UARTLCR_H_PEN
	}
	if parity == ParityEven {
		lcr |= rp.UART0_UARTLCR_H_EPS
	}
	const mask = rp.UART0_UARTLCR_H_WLEN_Msk | rp.UART0_UARTLCR_H_STP2 |
		rp.UART0_UARTLCR_H_PEN | rp.UART0_UARTLCR_H_EPS | rp.UART0_UARTLCR_H_SPS
	uart.Bus.UARTLCR_H.ReplaceBits(lcr, mask, 0)

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
	resetBlock(resetVal)
	unresetBlockWait(resetVal)
}

// handleInterrupt should be called from the appropriate interrupt handler for
// this UART instance.
func (uart *UART) handleInterrupt(interrupt.Interrupt) {
	for !uart.Bus.UARTFR.HasBits(rp.UART0_UARTFR_RXFE) {
		uart.Receive(byte((uart.Bus.UARTDR.Get() & 0xFF)))
	}
}
