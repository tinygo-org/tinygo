//go:build ch32v003

package machine

import "device/ch32"

const deviceName = ch32.Device

const (
	PA1 Pin = 1
	PA2 Pin = 2

	PC0 Pin = 16 + 0
	PC1 Pin = 16 + 1
	PC2 Pin = 16 + 2
	PC3 Pin = 16 + 3
	PC4 Pin = 16 + 4
	PC5 Pin = 16 + 5
	PC6 Pin = 16 + 6
	PC7 Pin = 16 + 7

	PD0 Pin = 24 + 0
	PD1 Pin = 24 + 1
	PD2 Pin = 24 + 2
	PD3 Pin = 24 + 3
	PD4 Pin = 24 + 4
	PD5 Pin = 24 + 5
	PD6 Pin = 24 + 6
	PD7 Pin = 24 + 7
)

const (
	PinInputAnalog         PinMode = 0b0000
	PinInputFloating       PinMode = 0b0100
	PinInputPullUpPullDown PinMode = 0b1000
	PinOutputPushPull      PinMode = 0b0001
	PinOutputOpenDrain     PinMode = 0b0101

	PinInput  = PinInputAnalog
	PinOutput = PinOutputPushPull
)

func (p Pin) getPortPin() (*ch32.GPIO_Type, uint32) {
	pin := uint32(p) & 0b111

	if p >= PD0 {
		return ch32.GPIOD, pin
	} else if p >= PC0 {
		return ch32.GPIOC, pin
	} else {
		return ch32.GPIOA, pin
	}
}

func (p Pin) Configure(config PinConfig) {
	mode := config.Mode
	port, pin := p.getPortPin()
	port.CFGLR.ReplaceBits(uint32(mode), 0b1111, uint8(pin)*4)
}

func (p Pin) Set(high bool) {
	port, pin := p.getPortPin()
	if high {
		port.BSHR.SetBits(1 << pin)
	} else {
		port.BCR.SetBits(1 << pin)
	}
}
