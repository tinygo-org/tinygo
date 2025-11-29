//go:build py32

package machine

import (
	"device/py32"
	"unsafe"
)

const deviceName = py32.Device

// Peripheral port offsets. Keep the same spacing used on other MCUs so helpers
// like Pin.getPort can keep using simple division by 16 even though PY32 ports
// only expose 8 pins each.
const (
	portA Pin = iota * 16
	portB
	portC
)

// Port A pins.
const (
	PA0 Pin = portA + iota
	PA1
	PA2
	PA3
	PA4
	PA5
	PA6
	PA7
)

// Port B pins.
const (
	PB0 Pin = portB + iota
	PB1
	PB2
	PB3
	PB4
	PB5
	PB6
	PB7
)

// Port C pins.
const (
	PC0 Pin = portC + iota
	PC1
	PC2
	PC3
	PC4
	PC5
	PC6
	PC7
)

// PinMode values specific to PY32: only GPIO direction and pull configuration.
const (
	PinOutput PinMode = iota
	PinInputFloating
	PinInputPulldown
	PinInputPullup
	PinInputAnalog
)
const PinInput PinMode = PinInputFloating

// Internal helpers for GPIO configuration.
const (
	gpioModeInput  = 0
	gpioModeOutput = 1
	gpioModeAnalog = 3
	gpioModeMask   = 0x3

	gpioPullFloating = 0
	gpioPullUp       = 1
	gpioPullDown     = 2
	gpioPullMask     = 0x3

	gpioOutputSpeedHigh = 2
	gpioOutputSpeedMask = 0x3
)

// // CPUFrequency returns the core clock frequency.
// func CPUFrequency() uint32 {
// 	return 48_000_000
// }

func (p Pin) getPortNumber() uint8 {
	return uint8(p) >> 4

}

func (p Pin) getPinNumber() uint8 {
	return uint8(p) & 0x0F
}

func (p Pin) getPort() (*py32.GPIO_Type, uint8) {
	offset := uintptr(p.getPortNumber()) * (uintptr(unsafe.Pointer(py32.GPIOB)) - uintptr(unsafe.Pointer(py32.GPIOA)))
	return (*py32.GPIO_Type)(unsafe.Pointer(uintptr(unsafe.Pointer(py32.GPIOA)) + offset)), p.getPinNumber()
}

func (p Pin) Set(high bool) {
	port, pin := p.getPort()
	if high {
		port.BSRR.Set(1 << pin)
	} else {
		port.BSRR.Set(1 << (pin + 16))
	}
}

func (p Pin) Get() bool {
	port, pin := p.getPort()
	val := port.IDR.Get() & (1 << pin)
	return val > 0
}

func (p Pin) Configure(config PinConfig) {
	p.enableClock()
	port, pin := p.getPort()
	pos := (pin % 16) * 2

	switch config.Mode {

	case PinInputFloating:
		port.MODER.ReplaceBits(gpioModeInput, gpioModeMask, pos)
		port.PUPDR.ReplaceBits(gpioPullFloating, gpioPullMask, pos)
	case PinInputPulldown:
		port.MODER.ReplaceBits(gpioModeInput, gpioModeMask, pos)
		port.PUPDR.ReplaceBits(gpioPullDown, gpioPullMask, pos)
	case PinInputPullup:
		port.MODER.ReplaceBits(gpioModeInput, gpioModeMask, pos)
		port.PUPDR.ReplaceBits(gpioPullUp, gpioPullMask, pos)
	case PinOutput:
		port.MODER.ReplaceBits(gpioModeOutput, gpioModeMask, pos)
		port.OTYPER.ReplaceBits(py32.GPIO_OTYPER_OT0_PushPull, py32.GPIO_OTYPER_OT0_Msk, pos>>1)
		port.OSPEEDR.ReplaceBits(gpioOutputSpeedHigh, gpioOutputSpeedMask, pos)
	case PinInputAnalog:
		port.MODER.ReplaceBits(gpioModeAnalog, gpioModeMask, pos)
		port.PUPDR.ReplaceBits(gpioPullFloating, gpioPullMask, pos)
	}
}

func (p Pin) SetAltFunc(af uint8) {
	port, pin := p.getPort()
	if pin >= 8 {
		port.AFRH.ReplaceBits(uint32(af), 0xF, (pin%8)*4)
	} else {
		port.AFRL.ReplaceBits(uint32(af), 0xF, (pin%8)*4)
	}
}

func (p Pin) enableClock() {
	portNo := p.getPortNumber()
	py32.RCC.IOPENR.SetBits(1 << portNo)
}
