//go:build ch32v003

package machine

import "device/wch"

const deviceName = wch.Device

const (
	PA0 Pin = 0 + 0
	PA1 Pin = 0 + 1
	PA2 Pin = 0 + 2
	PA3 Pin = 0 + 3
	PA4 Pin = 0 + 4
	PA5 Pin = 0 + 5
	PA6 Pin = 0 + 6
	PA7 Pin = 0 + 7
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
	PinInput  PinMode = 0
	PinOutput PinMode = 3
)

func (p Pin) getPortPin() (port *wch.GPIO_Type, pin uint32) {
	portNum := uint32(p) >> 3
	pin = uint32(p) & 0b111
	switch portNum {
	case 0:
		port = wch.GPIOA
	case 2:
		port = wch.GPIOC
	case 3:
		port = wch.GPIOD
	}
	return
}

func (p Pin) Configure(config PinConfig) {
	port, pin := p.getPortPin()
	reg := port.CFGLR.Get()
	reg &^= (0xf << (4 * pin))
	reg |= 0b0000 << (4 * pin)
	reg |= 0b11 << (4 * pin)
	port.CFGLR.Set(reg)
}

// Set the pin to high or low.
func (p Pin) Set(high bool) {
	port, pin := p.getPortPin()
	if high {
		port.BSHR.Set(1 << pin)
	} else {
		port.BCR.Set(1 << pin)
	}
}
