//go:build nrf52
// +build nrf52

package machine

import (
	"tinygo.org/x/device/nrf"
)

// Hardware pins
const (
	P0_00 Pin = 0
	P0_01 Pin = 1
	P0_02 Pin = 2
	P0_03 Pin = 3
	P0_04 Pin = 4
	P0_05 Pin = 5
	P0_06 Pin = 6
	P0_07 Pin = 7
	P0_08 Pin = 8
	P0_09 Pin = 9
	P0_10 Pin = 10
	P0_11 Pin = 11
	P0_12 Pin = 12
	P0_13 Pin = 13
	P0_14 Pin = 14
	P0_15 Pin = 15
	P0_16 Pin = 16
	P0_17 Pin = 17
	P0_18 Pin = 18
	P0_19 Pin = 19
	P0_20 Pin = 20
	P0_21 Pin = 21
	P0_22 Pin = 22
	P0_23 Pin = 23
	P0_24 Pin = 24
	P0_25 Pin = 25
	P0_26 Pin = 26
	P0_27 Pin = 27
	P0_28 Pin = 28
	P0_29 Pin = 29
	P0_30 Pin = 30
	P0_31 Pin = 31
)

// Get peripheral and pin number for this GPIO pin.
func (p Pin) getPortPin() (*nrf.GPIO_Type, uint32) {
	return nrf.P0, uint32(p)
}

func (uart *UART) setPins(tx, rx Pin) {
	nrf.UART0.PSELTXD.Set(uint32(tx))
	nrf.UART0.PSELRXD.Set(uint32(rx))
}

func (i2c *I2C) setPins(scl, sda Pin) {
	i2c.Bus.PSELSCL.Set(uint32(scl))
	i2c.Bus.PSELSDA.Set(uint32(sda))
}

// PWM
var (
	PWM0 = &PWM{PWM: nrf.PWM0}
	PWM1 = &PWM{PWM: nrf.PWM1}
	PWM2 = &PWM{PWM: nrf.PWM2}
)
