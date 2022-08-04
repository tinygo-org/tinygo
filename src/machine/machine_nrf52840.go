//go:build nrf52840
// +build nrf52840

package machine

import (
	"tinygo.org/x/device/nrf"
)

// Get peripheral and pin number for this GPIO pin.
func (p Pin) getPortPin() (*nrf.GPIO_Type, uint32) {
	if p >= 32 {
		return nrf.P1, uint32(p - 32)
	} else {
		return nrf.P0, uint32(p)
	}
}

func (uart *UART) setPins(tx, rx Pin) {
	nrf.UART0.PSEL.TXD.Set(uint32(tx))
	nrf.UART0.PSEL.RXD.Set(uint32(rx))
}

func (i2c *I2C) setPins(scl, sda Pin) {
	i2c.Bus.PSEL.SCL.Set(uint32(scl))
	i2c.Bus.PSEL.SDA.Set(uint32(sda))
}

// PWM
var (
	PWM0 = &PWM{PWM: nrf.PWM0}
	PWM1 = &PWM{PWM: nrf.PWM1}
	PWM2 = &PWM{PWM: nrf.PWM2}
	PWM3 = &PWM{PWM: nrf.PWM3}
)
