// +build esp32

package machine

import (
	"device/esp"
	"runtime/volatile"
)

const peripheralClock = 80000000 // 80MHz

// CPUFrequency returns the current CPU frequency of the chip.
// Currently it is a fixed frequency but it may allow changing in the future.
func CPUFrequency() uint32 {
	return 160e6 // 160MHz
}

type PinMode uint8

const (
	PinOutput PinMode = iota
	PinInput
	PinInputPullup
	PinInputPulldown
)

// Configure this pin with the given configuration.
func (p Pin) Configure(config PinConfig) {
	var muxConfig uint32 // The mux configuration.

	// Configure this pin as a GPIO pin.
	const function = 3 // function 3 is GPIO for every pin
	muxConfig |= (function - 1) << esp.IO_MUX_GPIO0_MCU_SEL_Pos

	// Make this pin an input pin (always).
	muxConfig |= esp.IO_MUX_GPIO0_FUN_IE

	// Set drive strength: 0 is lowest, 3 is highest.
	muxConfig |= 2 << esp.IO_MUX_GPIO0_FUN_DRV_Pos

	// Select pull mode.
	if config.Mode == PinInputPullup {
		muxConfig |= esp.IO_MUX_GPIO0_FUN_WPU
	} else if config.Mode == PinInputPulldown {
		muxConfig |= esp.IO_MUX_GPIO0_FUN_WPD
	}

	// Configure the pad with the given IO mux configuration.
	p.mux().Set(muxConfig)

	switch config.Mode {
	case PinOutput:
		// Set the 'output enable' bit.
		if p < 32 {
			esp.GPIO.ENABLE_W1TS.Set(1 << p)
		} else {
			esp.GPIO.ENABLE1_W1TS.Set(1 << (p - 32))
		}
	case PinInput, PinInputPullup, PinInputPulldown:
		// Clear the 'output enable' bit.
		if p < 32 {
			esp.GPIO.ENABLE_W1TC.Set(1 << p)
		} else {
			esp.GPIO.ENABLE1_W1TC.Set(1 << (p - 32))
		}
	}
}

// Set the pin to high or low.
// Warning: only use this on an output pin!
func (p Pin) Set(value bool) {
	if value {
		reg, mask := p.portMaskSet()
		reg.Set(mask)
	} else {
		reg, mask := p.portMaskClear()
		reg.Set(mask)
	}
}

// Return the register and mask to enable a given GPIO pin. This can be used to
// implement bit-banged drivers.
//
// Warning: only use this on an output pin!
func (p Pin) PortMaskSet() (*uint32, uint32) {
	reg, mask := p.portMaskSet()
	return &reg.Reg, mask
}

// Return the register and mask to disable a given GPIO pin. This can be used to
// implement bit-banged drivers.
//
// Warning: only use this on an output pin!
func (p Pin) PortMaskClear() (*uint32, uint32) {
	reg, mask := p.portMaskClear()
	return &reg.Reg, mask
}

func (p Pin) portMaskSet() (*volatile.Register32, uint32) {
	if p < 32 {
		return &esp.GPIO.OUT_W1TS, 1 << p
	} else {
		return &esp.GPIO.OUT1_W1TS, 1 << (p - 32)
	}
}

func (p Pin) portMaskClear() (*volatile.Register32, uint32) {
	if p < 32 {
		return &esp.GPIO.OUT_W1TC, 1 << p
	} else {
		return &esp.GPIO.OUT1_W1TC, 1 << (p - 32)
	}
}

// Get returns the current value of a GPIO pin when the pin is configured as an
// input.
func (p Pin) Get() bool {
	if p < 32 {
		return esp.GPIO.IN.Get()&(1<<p) != 0
	} else {
		return esp.GPIO.IN1.Get()&(1<<(p-32)) != 0
	}
}

// mux returns the I/O mux configuration register corresponding to the given
// GPIO pin.
func (p Pin) mux() *volatile.Register32 {
	// I have no idea whether there is any pattern in the GPIO <-> pad mapping.
	// I couldn't find it.
	switch p {
	case 36:
		return &esp.IO_MUX.GPIO36
	case 37:
		return &esp.IO_MUX.GPIO37
	case 38:
		return &esp.IO_MUX.GPIO38
	case 39:
		return &esp.IO_MUX.GPIO39
	case 34:
		return &esp.IO_MUX.GPIO34
	case 35:
		return &esp.IO_MUX.GPIO35
	case 32:
		return &esp.IO_MUX.GPIO32
	case 33:
		return &esp.IO_MUX.GPIO33
	case 25:
		return &esp.IO_MUX.GPIO25
	case 26:
		return &esp.IO_MUX.GPIO26
	case 27:
		return &esp.IO_MUX.GPIO27
	case 14:
		return &esp.IO_MUX.MTMS
	case 12:
		return &esp.IO_MUX.MTDI
	case 13:
		return &esp.IO_MUX.MTCK
	case 15:
		return &esp.IO_MUX.MTDO
	case 2:
		return &esp.IO_MUX.GPIO2
	case 0:
		return &esp.IO_MUX.GPIO0
	case 4:
		return &esp.IO_MUX.GPIO4
	case 16:
		return &esp.IO_MUX.GPIO16
	case 17:
		return &esp.IO_MUX.GPIO17
	case 9:
		return &esp.IO_MUX.SD_DATA2
	case 10:
		return &esp.IO_MUX.SD_DATA3
	case 11:
		return &esp.IO_MUX.SD_CMD
	case 6:
		return &esp.IO_MUX.SD_CLK
	case 7:
		return &esp.IO_MUX.SD_DATA0
	case 8:
		return &esp.IO_MUX.SD_DATA1
	case 5:
		return &esp.IO_MUX.GPIO5
	case 18:
		return &esp.IO_MUX.GPIO18
	case 19:
		return &esp.IO_MUX.GPIO19
	case 20:
		return &esp.IO_MUX.GPIO20
	case 21:
		return &esp.IO_MUX.GPIO21
	case 22:
		return &esp.IO_MUX.GPIO22
	case 3:
		return &esp.IO_MUX.U0RXD
	case 1:
		return &esp.IO_MUX.U0TXD
	case 23:
		return &esp.IO_MUX.GPIO23
	case 24:
		return &esp.IO_MUX.GPIO24
	default:
		return nil
	}
}

var (
	UART0 = UART{Bus: esp.UART0, Buffer: NewRingBuffer()}
	UART1 = UART{Bus: esp.UART1, Buffer: NewRingBuffer()}
	UART2 = UART{Bus: esp.UART2, Buffer: NewRingBuffer()}
)

type UART struct {
	Bus    *esp.UART_Type
	Buffer *RingBuffer
}

func (uart UART) Configure(config UARTConfig) {
	if config.BaudRate == 0 {
		config.BaudRate = 115200
	}
	uart.Bus.CLKDIV.Set(peripheralClock / config.BaudRate)
}

func (uart UART) WriteByte(b byte) error {
	for (uart.Bus.STATUS.Get()>>16)&0xff >= 128 {
		// Read UART_TXFIFO_CNT from the status register, which indicates how
		// many bytes there are in the transmit buffer. Wait until there are
		// less than 128 bytes in this buffer (the default buffer size).
	}
	uart.Bus.TX_FIFO.Set(b)
	return nil
}
