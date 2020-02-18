// +build nxp,mk66f18

package machine

import (
	"device/arm"
	"device/nxp"
	"errors"
	"runtime/volatile"

	_ "unsafe" // for go:linkname
)

const (
	uartC2Enable       = nxp.UART_C2_TE | nxp.UART_C2_RE | nxp.UART_C2_RIE | nxp.UART_C2_ILIE
	uartC2TXActive     = uartC2Enable | nxp.UART_C2_TIE
	uartC2TXCompleting = uartC2Enable | nxp.UART_C2_TCIE
	uartC2TXInactive   = uartC2Enable

	uartIRQPriority = 64
)

var (
	ErrNotImplemented = errors.New("device has not been implemented")
	ErrNotConfigured  = errors.New("device has not been configured")
)

type UARTConfig struct {
	BaudRate uint32
}

type UART struct {
	*nxp.UART_Type
	RXPCR     *volatile.Register32
	TXPCR     *volatile.Register32
	SCGC      *volatile.Register32
	SCGCMask  uint32
	IRQNumber uint32

	// state
	RXBuffer     RingBuffer
	TXBuffer     RingBuffer
	Transmitting volatile.Register8
}

// 'UART0' in the K66 manual corresponds to 'UART1' on the Teensy's pinout
var UART1 = UART{UART_Type: nxp.UART0, RXPCR: pins[0].PCR, TXPCR: pins[1].PCR, SCGC: &nxp.SIM.SCGC4, SCGCMask: nxp.SIM_SCGC4_UART0, IRQNumber: nxp.IRQ_UART0_RX_TX}
var UART2 = UART{UART_Type: nxp.UART1, RXPCR: pins[9].PCR, TXPCR: pins[10].PCR, SCGC: &nxp.SIM.SCGC4, SCGCMask: nxp.SIM_SCGC4_UART1, IRQNumber: nxp.IRQ_UART1_RX_TX}
var UART3 = UART{UART_Type: nxp.UART2, RXPCR: pins[7].PCR, TXPCR: pins[8].PCR, SCGC: &nxp.SIM.SCGC4, SCGCMask: nxp.SIM_SCGC4_UART2, IRQNumber: nxp.IRQ_UART2_RX_TX}
var UART4 = UART{UART_Type: nxp.UART3, RXPCR: pins[31].PCR, TXPCR: pins[32].PCR, SCGC: &nxp.SIM.SCGC4, SCGCMask: nxp.SIM_SCGC4_UART3, IRQNumber: nxp.IRQ_UART3_RX_TX}
var UART5 = UART{UART_Type: nxp.UART4, RXPCR: pins[34].PCR, TXPCR: pins[33].PCR, SCGC: &nxp.SIM.SCGC1, SCGCMask: nxp.SIM_SCGC1_UART4, IRQNumber: nxp.IRQ_UART4_RX_TX}

//go:export UART0_RX_TX_IRQHandler
func uart0StatusISR() { UART1.handleStatusInterrupt() }

//go:export UART1_RX_TX_IRQHandler
func uart1StatusISR() { UART2.handleStatusInterrupt() }

//go:export UART2_RX_TX_IRQHandler
func uart2StatusISR() { UART3.handleStatusInterrupt() }

//go:export UART3_RX_TX_IRQHandler
func uart3StatusISR() { UART4.handleStatusInterrupt() }

//go:export UART4_RX_TX_IRQHandler
func uart4StatusISR() { UART5.handleStatusInterrupt() }

// Configure the UART.
func (u *UART) Configure(config UARTConfig) {
	en := u.SCGC.HasBits(u.SCGCMask)

	// adapted from Teensy core's serial_begin

	if !en {
		u.Transmitting.Set(0)

		// turn on the clock
		u.SCGC.Set(u.SCGCMask)

		// configure pins
		u.RXPCR.Set(nxp.PORT_PCR0_PE | nxp.PORT_PCR0_PS | nxp.PORT_PCR0_PFE | nxp.PORT_PCR0_MUX(3))
		u.TXPCR.Set(nxp.PORT_PCR0_DSE | nxp.PORT_PCR0_SRE | nxp.PORT_PCR0_MUX(3))
		u.C1.Set(nxp.UART_C1_ILT)
	}

	// default to 115200 baud
	if config.BaudRate == 0 {
		config.BaudRate = 115200
	}

	// copied from teensy core's BAUD2DIV macro
	divisor := ((CPUFrequency() * 2) + ((config.BaudRate) >> 1)) / config.BaudRate
	if divisor < 32 {
		divisor = 32
	}

	if en {
		// don't change baud rate mid transmit
		u.Flush()
	}

	// set the divisor
	u.BDH.Set(uint8((divisor >> 13) & 0x1F))
	u.BDL.Set(uint8((divisor >> 5) & 0xFF))
	u.C4.Set(uint8(divisor & 0x1F))

	if !en {
		u.C1.Set(nxp.UART_C1_ILT)

		// configure TX and RX watermark
		u.TWFIFO.Set(2) // causes bit TDRE of S1 to set
		u.RWFIFO.Set(4) // causes bit RDRF of S1 to set

		u.PFIFO.Set(nxp.UART_PFIFO_TXFE | nxp.UART_PFIFO_RXFE)
		u.C2.Set(uartC2TXInactive)

		arm.SetPriority(u.IRQNumber, uartIRQPriority)
		arm.EnableIRQ(u.IRQNumber)
	}
}

func (u *UART) Disable() {
	// adapted from Teensy core's serial_end

	// check if the device has been enabled already
	if !u.SCGC.HasBits(u.SCGCMask) {
		return
	}

	u.Flush()

	arm.DisableIRQ(u.IRQNumber)
	u.C2.Set(0)

	// reconfigure pin
	u.RXPCR.Set(nxp.PORT_PCR0_PE | nxp.PORT_PCR0_PS | nxp.PORT_PCR0_MUX(1))
	u.TXPCR.Set(nxp.PORT_PCR0_PE | nxp.PORT_PCR0_PS | nxp.PORT_PCR0_MUX(1))

	// clear flags
	u.S1.Get()
	u.D.Get()
	u.RXBuffer.Clear()
}

func (u *UART) Flush() {
	for u.Transmitting.Get() != 0 {
		// gosched()
	}
}

// adapted from Teensy core's uart0_status_isr
func (u *UART) handleStatusInterrupt() {
	if u.S1.HasBits(nxp.UART_S1_RDRF | nxp.UART_S1_IDLE) {
		intrs := arm.DisableInterrupts()
		avail := u.RCFIFO.Get()
		if avail == 0 {
			// The only way to clear the IDLE interrupt flag is
			// to read the data register.  But reading with no
			// data causes a FIFO underrun, which causes the
			// FIFO to return corrupted data.  If anyone from
			// Freescale reads this, what a poor design!  There
			// write should be a write-1-to-clear for IDLE.
			u.D.Get()
			// flushing the fifo recovers from the underrun,
			// but there's a possible race condition where a
			// new character could be received between reading
			// RCFIFO == 0 and flushing the FIFO.  To minimize
			// the chance, interrupts are disabled so a higher
			// priority interrupt (hopefully) doesn't delay.
			// TODO: change this to disabling the IDLE interrupt
			// which won't be simple, since we already manage
			// which transmit interrupts are enabled.
			u.CFIFO.Set(nxp.UART_CFIFO_RXFLUSH)
			arm.EnableInterrupts(intrs)

		} else {
			arm.EnableInterrupts(intrs)

			for {
				u.RXBuffer.Put(u.D.Get())
				avail--
				if avail <= 0 {
					break
				}
			}
		}
	}

	c := u.C2.Get()
	if c&nxp.UART_C2_TIE != 0 && u.S1.HasBits(nxp.UART_S1_TDRE) {
		for {
			n, ok := u.TXBuffer.Get()
			if !ok {
				break
			}

			u.S1.Get()
			u.D.Set(n)

			if u.TCFIFO.Get() >= 8 {
				break
			}
		}

		if u.S1.HasBits(nxp.UART_S1_TDRE) {
			u.Transmitting.Set(0)
			u.C2.Set(uartC2TXCompleting)
		}
	}

	if c&nxp.UART_C2_TCIE != 0 && u.S1.HasBits(nxp.UART_S1_TC) {
		u.C2.Set(uartC2TXInactive)
	}
}

//go:linkname gosched runtime.Gosched
func gosched()

// WriteByte writes a byte of data to the UART.
func (u *UART) WriteByte(c byte) error {
	if !u.SCGC.HasBits(u.SCGCMask) {
		return ErrNotConfigured
	}

	for !u.S1.HasBits(nxp.UART_S1_TDRE) {
		gosched()
	}
	u.D.Set(c)

	// // wait for room on the buffer
	// for !u.TXBuffer.Put(c) {
	// 	gosched()
	// }

	// var wrote bool
	// for u.S1.HasBits(nxp.UART_S1_TDRE) {
	// 	n, ok := u.TXBuffer.Get()
	// 	if ok {
	// 		u.D.Set(n)
	// 		wrote = true
	// 	} else {
	// 		break
	// 	}
	// }

	// if wrote {
	// 	u.Transmitting.Set(1)
	// 	u.C2.Set(uartC2TXActive)
	// }
	return nil
}
