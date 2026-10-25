//go:build sam && atsam3x8e

package machine

import (
	"device/sam"
	"runtime/interrupt"
)

const deviceName = sam.Device

// CPUFrequency returns the MCK frequency set by the runtime.
func CPUFrequency() uint32 {
	return 84_000_000
}

// PA0 through PD10 identify the SAM3X8E port pins.
const (
	PA0  Pin = 0
	PA1  Pin = 1
	PA2  Pin = 2
	PA3  Pin = 3
	PA4  Pin = 4
	PA5  Pin = 5
	PA6  Pin = 6
	PA7  Pin = 7
	PA8  Pin = 8
	PA9  Pin = 9
	PA10 Pin = 10
	PA11 Pin = 11
	PA12 Pin = 12
	PA13 Pin = 13
	PA14 Pin = 14
	PA15 Pin = 15
	PA16 Pin = 16
	PA17 Pin = 17
	PA18 Pin = 18
	PA19 Pin = 19
	PA20 Pin = 20
	PA21 Pin = 21
	PA22 Pin = 22
	PA23 Pin = 23
	PA24 Pin = 24
	PA25 Pin = 25
	PA26 Pin = 26
	PA27 Pin = 27
	PA28 Pin = 28
	PA29 Pin = 29
	PB0  Pin = 32
	PB1  Pin = 33
	PB2  Pin = 34
	PB3  Pin = 35
	PB4  Pin = 36
	PB5  Pin = 37
	PB6  Pin = 38
	PB7  Pin = 39
	PB8  Pin = 40
	PB9  Pin = 41
	PB10 Pin = 42
	PB11 Pin = 43
	PB12 Pin = 44
	PB13 Pin = 45
	PB14 Pin = 46
	PB15 Pin = 47
	PB16 Pin = 48
	PB17 Pin = 49
	PB18 Pin = 50
	PB19 Pin = 51
	PB20 Pin = 52
	PB21 Pin = 53
	PB22 Pin = 54
	PB23 Pin = 55
	PB24 Pin = 56
	PB25 Pin = 57
	PB26 Pin = 58
	PB27 Pin = 59
	PB28 Pin = 60
	PB29 Pin = 61
	PB30 Pin = 62
	PB31 Pin = 63
	PC0  Pin = 64
	PC1  Pin = 65
	PC2  Pin = 66
	PC3  Pin = 67
	PC4  Pin = 68
	PC5  Pin = 69
	PC6  Pin = 70
	PC7  Pin = 71
	PC8  Pin = 72
	PC9  Pin = 73
	PC10 Pin = 74
	PC11 Pin = 75
	PC12 Pin = 76
	PC13 Pin = 77
	PC14 Pin = 78
	PC15 Pin = 79
	PC16 Pin = 80
	PC17 Pin = 81
	PC18 Pin = 82
	PC19 Pin = 83
	PC20 Pin = 84
	PC21 Pin = 85
	PC22 Pin = 86
	PC23 Pin = 87
	PC24 Pin = 88
	PC25 Pin = 89
	PC26 Pin = 90
	PC27 Pin = 91
	PC28 Pin = 92
	PC29 Pin = 93
	PC30 Pin = 94
	PD0  Pin = 96
	PD1  Pin = 97
	PD2  Pin = 98
	PD3  Pin = 99
	PD4  Pin = 100
	PD5  Pin = 101
	PD6  Pin = 102
	PD7  Pin = 103
	PD8  Pin = 104
	PD9  Pin = 105
	PD10 Pin = 106
)

const (
	PinOutput PinMode = iota
	PinInput
	PinInputPullup
	PinInputPulldown
	PinPeripheralA
	PinPeripheralB
)

func (p Pin) port() *sam.PIO_Type {
	switch p / 32 {
	case 0:
		return sam.PIOA
	case 1:
		return sam.PIOB
	case 2:
		return sam.PIOC
	default:
		return sam.PIOD
	}
}

func (p Pin) mask() uint32 {
	return 1 << (p % 32)
}

// Configure sets the pin mode.
// PinInputPulldown acts as PinInput because PIO has no pull-down.
func (p Pin) Configure(config PinConfig) {
	if p == NoPin {
		return
	}
	enablePIOClocks()
	port, mask := p.port(), p.mask()

	switch config.Mode {
	case PinOutput:
		port.IDR.Set(mask)
		port.PUDR.Set(mask)
		port.OER.Set(mask)
		port.PER.Set(mask)
	case PinInputPullup:
		port.IDR.Set(mask)
		port.PUER.Set(mask)
		port.ODR.Set(mask)
		port.PER.Set(mask)
	case PinPeripheralA, PinPeripheralB:
		port.IDR.Set(mask)
		if config.Mode == PinPeripheralA {
			port.ABSR.ClearBits(mask)
		} else {
			port.ABSR.SetBits(mask)
		}
		port.PDR.Set(mask)
	default:
		port.IDR.Set(mask)
		port.PUDR.Set(mask)
		port.ODR.Set(mask)
		port.PER.Set(mask)
	}
}

// Set drives the pin high if value is true, or low if value is false.
func (p Pin) Set(value bool) {
	if p == NoPin {
		return
	}
	if value {
		p.port().SODR.Set(p.mask())
	} else {
		p.port().CODR.Set(p.mask())
	}
}

// Get reports whether the pin input is high.
func (p Pin) Get() bool {
	if p == NoPin {
		return false
	}
	return p.port().PDSR.HasBits(p.mask())
}

var pioClocksEnabled bool

// enablePIOClocks turns on the PIO clocks so that PDSR reads the pins.
func enablePIOClocks() {
	if pioClocksEnabled {
		return
	}
	sam.PMC.PMC_PCER0.Set(1<<sam.IRQ_PIOA | 1<<sam.IRQ_PIOB | 1<<sam.IRQ_PIOC | 1<<sam.IRQ_PIOD)
	pioClocksEnabled = true
}

// UART on the SAM3X8E. Only the UART peripheral is supported, not the USARTs.
type UART struct {
	Buffer    *RingBuffer
	Bus       *sam.UART_Type
	Interrupt interrupt.Interrupt
}

var (
	DefaultUART = UART0
	UART0       = &_UART0
	_UART0      = UART{Buffer: NewRingBuffer(), Bus: sam.UART}
)

func init() {
	UART0.Interrupt = interrupt.New(sam.IRQ_UART, _UART0.handleInterrupt)
}

// Configure the UART. The UART peripheral is fixed to PA8 and PA9.
func (uart *UART) Configure(config UARTConfig) error {
	if config.BaudRate == 0 {
		config.BaudRate = 115200
	}

	PA8.Configure(PinConfig{Mode: PinPeripheralA})
	PA9.Configure(PinConfig{Mode: PinPeripheralA})
	sam.PMC.PMC_PCER0.Set(1 << sam.IRQ_UART)

	uart.Bus.CR.Set(sam.UART_CR_RSTRX | sam.UART_CR_RSTTX | sam.UART_CR_RXDIS | sam.UART_CR_TXDIS)
	uart.Bus.MR.Set(sam.UART_MR_PAR_NO << sam.UART_MR_PAR_Pos)
	uart.SetBaudRate(config.BaudRate)
	uart.Bus.IDR.Set(0xffffffff)
	uart.Bus.IER.Set(sam.UART_IER_RXRDY)
	uart.Bus.CR.Set(sam.UART_CR_RXEN | sam.UART_CR_TXEN)

	uart.Interrupt.Enable()
	return nil
}

// SetBaudRate sets the UART baud rate.
func (uart *UART) SetBaudRate(br uint32) {
	uart.Bus.BRGR.Set((CPUFrequency() + br*8) / (br * 16))
}

func (uart *UART) writeByte(c byte) error {
	for !uart.Bus.SR.HasBits(sam.UART_SR_TXRDY) {
	}
	uart.Bus.THR.Set(uint32(c))
	return nil
}

func (uart *UART) flush() {
	for !uart.Bus.SR.HasBits(sam.UART_SR_TXEMPTY) {
	}
}

func (uart *UART) handleInterrupt(interrupt.Interrupt) {
	if uart.Bus.SR.HasBits(sam.UART_SR_RXRDY) {
		uart.Receive(byte(uart.Bus.RHR.Get()))
	}
}
