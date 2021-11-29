// +build esp32c3

package machine

import (
	"device/esp"
	"runtime/volatile"
	"unsafe"
)

const deviceName = esp.Device

// CPUFrequency returns the current CPU frequency of the chip.
// Currently it is a fixed frequency but it may allow changing in the future.
func CPUFrequency() uint32 {
	return 160e6 // 160MHz
}

const (
	PinOutput PinMode = iota
	PinInput
	PinInputPullup
	PinInputPulldown
)

// Configure this pin with the given configuration.
func (p Pin) Configure(config PinConfig) {
	if p == NoPin {
		// This simplifies pin configuration in peripherals such as SPI.
		return
	}

	var muxConfig uint32

	// Configure this pin as a GPIO pin.
	const function = 1 // function 1 is GPIO for every pin
	muxConfig |= function << esp.IO_MUX_GPIO_MCU_SEL_Pos

	// Make this pin an input pin (always).
	muxConfig |= esp.IO_MUX_GPIO_FUN_IE

	// Set drive strength: 0 is lowest, 3 is highest.
	muxConfig |= 2 << esp.IO_MUX_GPIO_FUN_DRV_Pos

	// Select pull mode.
	if config.Mode == PinInputPullup {
		muxConfig |= esp.IO_MUX_GPIO_FUN_WPU
	} else if config.Mode == PinInputPulldown {
		muxConfig |= esp.IO_MUX_GPIO_FUN_WPD
	}

	// Configure the pad with the given IO mux configuration.
	p.mux().Set(muxConfig)

	// Set the output signal to the simple GPIO output.
	p.outFunc().Set(0x80)

	switch config.Mode {
	case PinOutput:
		// Set the 'output enable' bit.
		esp.GPIO.ENABLE_W1TS.Set(1 << p)
	case PinInput, PinInputPullup, PinInputPulldown:
		// Clear the 'output enable' bit.
		esp.GPIO.ENABLE_W1TC.Set(1 << p)
	}
}

// outFunc returns the FUNCx_OUT_SEL_CFG register used for configuring the
// output function selection.
func (p Pin) outFunc() *volatile.Register32 {
	return (*volatile.Register32)(unsafe.Pointer((uintptr(unsafe.Pointer(&esp.GPIO.FUNC0_OUT_SEL_CFG)) + uintptr(p)*4)))
}

// inFunc returns the FUNCy_IN_SEL_CFG register used for configuring the input
// function selection.
func inFunc(signal uint32) *volatile.Register32 {
	return (*volatile.Register32)(unsafe.Pointer((uintptr(unsafe.Pointer(&esp.GPIO.FUNC0_IN_SEL_CFG)) + uintptr(signal)*4)))
}

// mux returns the I/O mux configuration register corresponding to the given
// GPIO pin.
func (p Pin) mux() *volatile.Register32 {
	return (*volatile.Register32)(unsafe.Pointer((uintptr(unsafe.Pointer(&esp.IO_MUX.GPIO0)) + uintptr(p)*4)))
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
	return &esp.GPIO.OUT_W1TS, 1 << p
}

func (p Pin) portMaskClear() (*volatile.Register32, uint32) {
	return &esp.GPIO.OUT_W1TC, 1 << p
}

var DefaultUART = UART0

var (
	UART0  = &_UART0
	_UART0 = UART{Bus: esp.UART0, Buffer: NewRingBuffer()}
	UART1  = &_UART1
	_UART1 = UART{Bus: esp.UART1, Buffer: NewRingBuffer()}
)

type UART struct {
	Bus    *esp.UART_Type
	Buffer *RingBuffer
}

func (uart *UART) WriteByte(b byte) error {
	for (uart.Bus.STATUS.Get()&esp.UART_STATUS_TXFIFO_CNT_Msk)>>esp.UART_STATUS_TXFIFO_CNT_Pos >= 128 {
		// Read UART_TXFIFO_CNT from the status register, which indicates how
		// many bytes there are in the transmit buffer. Wait until there are
		// less than 128 bytes in this buffer (the default buffer size).
	}
	uart.Bus.FIFO.Set(uint32(b))
	return nil
}
