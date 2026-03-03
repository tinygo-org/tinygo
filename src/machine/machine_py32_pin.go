//go:build py32

package machine

import (
	"device/py32"
	"unsafe"
)

const deviceName = py32.Device

// Peripheral port offsets.
// Keep the same spacing used on other MCUs so helpers like Pin.getPortNumber
// can keep using simple division by 16.
const (
	portA Pin = iota * 16
	portB
	portC
	portD
	portE
	portF
)

const (
	PA0 Pin = portA + iota
	PA1
	PA2
	PA3
	PA4
	PA5
	PA6
	PA7
	PA8
	PA9
	PA10
	PA11
	PA12
	PA13
	PA14
	PA15
)

const (
	PB0 Pin = portB + iota
	PB1
	PB2
	PB3
	PB4
	PB5
	PB6
	PB7
	PB8
	PB9
	PB10
	PB11
	PB12
	PB13
	PB14
	PB15
)

const (
	PC0 Pin = portC + iota
	PC1
	PC2
	PC3
	PC4
	PC5
	PC6
	PC7
	PC8
	PC9
	PC10
	PC11
	PC12
	PC13
	PC14
	PC15
)

const (
	PD0 Pin = portD + iota
	PD1
	PD2
	PD3
	PD4
	PD5
	PD6
	PD7
	PD8
	PD9
	PD10
	PD11
	PD12
	PD13
	PD14
	PD15
)

const (
	PE0 Pin = portE + iota
	PE1
	PE2
	PE3
	PE4
	PE5
	PE6
	PE7
	PE8
	PE9
	PE10
	PE11
	PE12
	PE13
	PE14
	PE15
)

const (
	PF0 Pin = portF + iota
	PF1
	PF2
	PF3
	PF4
	PF5
	PF6
	PF7
	PF8
	PF9
	PF10
	PF11
	PF12
	PF13
	PF14
	PF15
)

const (
	PinOutput PinMode = iota
	PinInputFloating
	PinInputPulldown
	PinInputPullup
	PinInputAnalog
	PinAlternate
)
const PinInput PinMode = PinInputFloating

const (
	gpioModeInput     = py32.GPIO_MODER_MODE0_Input
	gpioModeOutput    = py32.GPIO_MODER_MODE0_Output
	gpioModeAlternate = py32.GPIO_MODER_MODE0_Alternate
	gpioModeAnalog    = py32.GPIO_MODER_MODE0_Analog
	gpioModeMask      = py32.GPIO_MODER_MODE0_Msk

	gpioPullFloating = py32.GPIO_PUPDR_PUPD0_Floating
	gpioPullUp       = py32.GPIO_PUPDR_PUPD0_PullUp
	gpioPullDown     = py32.GPIO_PUPDR_PUPD0_PullDown
	gpioPullMask     = py32.GPIO_PUPDR_PUPD0_Msk

	gpioOutputSpeedLow      = py32.GPIO_OSPEEDR_OSPEED0_LowSpeed
	gpioOutputSpeedMedium   = py32.GPIO_OSPEEDR_OSPEED0_MediumSpeed
	gpioOutputSpeedHigh     = py32.GPIO_OSPEEDR_OSPEED0_HighSpeed
	gpioOutputSpeedVeryHigh = py32.GPIO_OSPEEDR_OSPEED0_VeryHighSpeed
	gpioOutputSpeedMask     = py32.GPIO_OSPEEDR_OSPEED0_Msk
)

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
	case PinAlternate:
		port.MODER.ReplaceBits(gpioModeAlternate, gpioModeMask, pos)
		port.OSPEEDR.ReplaceBits(gpioOutputSpeedHigh, gpioOutputSpeedMask, pos)
	}
}

func (p Pin) enableClock() {
	portNo := p.getPortNumber()
	py32.RCC.IOPENR.SetBits(1 << portNo)
}
