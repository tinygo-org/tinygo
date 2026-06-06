//go:build stm32

package machine

// Peripheral abstraction layer for UARTs on the stm32 family (except stm32g0).

import (
	"device/stm32"
	"runtime/interrupt"
	"runtime/volatile"
	"unsafe"
)

// UART representation
type UART struct {
	Buffer            *RingBuffer
	Bus               *stm32.USART_Type
	Interrupt         interrupt.Interrupt
	TxAltFuncSelector uint8
	RxAltFuncSelector uint8

	// Registers specific to the chip
	rxReg       *volatile.Register32
	txReg       *volatile.Register32
	statusReg   *volatile.Register32
	txEmptyFlag uint32
	// errClearReg points to the ICR register on newer STM32 USART peripherals
	// (L0, L4, L5, G0, F7, U5, WL, etc.) for clearing error flags. Nil for
	// older peripherals (F1, F4) where errors are cleared by reading SR+DR.
	errClearReg *volatile.Register32
}

// Configure the UART.
func (uart *UART) Configure(config UARTConfig) {
	// Default baud rate to 115200.
	if config.BaudRate == 0 {
		config.BaudRate = 115200
	}

	// Set the GPIO pins to defaults if they're not set
	if config.TX == 0 && config.RX == 0 {
		config.TX = UART_TX_PIN
		config.RX = UART_RX_PIN
	}

	// STM32 families have different, but compatible, registers for
	// basic UART functions.  For each family populate the registers
	// into `uart`.
	uart.setRegisters()

	// Enable USART clock
	enableAltFuncClock(unsafe.Pointer(uart.Bus))

	uart.configurePins(config)

	// Set baud rate
	uart.SetBaudRate(config.BaudRate)

	// Enable USART port, tx, rx and rx interrupts
	uart.Bus.CR1.Set(stm32.USART_CR1_TE | stm32.USART_CR1_RE | stm32.USART_CR1_RXNEIE | stm32.USART_CR1_UE)

	// Enable RX IRQ
	uart.Interrupt.SetPriority(0xc0)
	uart.Interrupt.Enable()
}

// handleInterrupt should be called from the appropriate interrupt handler for
// this UART instance.
func (uart *UART) handleInterrupt(interrupt.Interrupt) {
	s := uart.statusReg.Get()

	// Only read data when RXNE/RXFNE (bit 5) is set. On all STM32 families,
	// RXNEIE enables both the RX-data-ready and overrun-error (ORE) interrupts.
	// Without this check, an ORE-only interrupt reads garbage from RDR.
	if s&0x20 != 0 { // RXNE / RXFNE
		uart.Receive(byte((uart.rxReg.Get() & 0xFF)))
	}

	// Clear error flags (ORE=bit3, NE=bit2, FE=bit1, PE=bit0) to prevent
	// an interrupt storm and ensure the USART can continue receiving.
	if s&0xF != 0 {
		if uart.errClearReg != nil {
			// Newer USART peripherals (L0, L4, L5, G0, F7, U5, WL, etc.):
			// clear all error flags via ICR (ORECF|NECF|FECF|PECF = bits 3:0).
			uart.errClearReg.Set(s & 0xF)
		} else if s&0x20 == 0 {
			// Older USART (F1/F4): errors are cleared by reading SR then DR.
			// SR was already read above. If RXNE was set, DR was read in
			// the Receive path. Otherwise do a dummy DR read to complete
			// the clearing sequence.
			uart.rxReg.Get()
		}
	}
}

// WriteByte writes a byte of data to the UART.
func (uart *UART) writeByte(c byte) error {
	// Wait for the transmit data register to be empty before writing, so we
	// don't overwrite a byte that hasn't moved to the shift register yet.
	for !uart.statusReg.HasBits(uart.txEmptyFlag) {
	}
	uart.txReg.Set(uint32(c))
	return nil
}

// flush waits until the USART shift register has finished transmitting the
// last byte (TC = Transmission Complete, bit 6). Without this, Write() returns
// while the final byte is still clocking out on the wire.
func (uart *UART) flush() {
	for !uart.statusReg.HasBits(1 << 6) { // TC bit
	}
}
