// +build nrf

package machine

import (
	"device/nrf"
)

type GPIOMode uint8

const (
	GPIO_INPUT  = nrf.P0_PIN_CNF_DIR_Input
	GPIO_OUTPUT = nrf.P0_PIN_CNF_DIR_Output
)

// LEDs on the PCA10040 (nRF52832 dev board)
const (
	LED  = LED1
	LED1 = 17
	LED2 = 18
	LED3 = 19
	LED4 = 20
)

func (p GPIO) Configure(config GPIOConfig) {
	cfg := config.Mode | nrf.P0_PIN_CNF_PULL_Disabled | nrf.P0_PIN_CNF_DRIVE_S0S1 | nrf.P0_PIN_CNF_SENSE_Disabled
	if config.Mode == GPIO_INPUT {
		cfg |= nrf.P0_PIN_CNF_INPUT_Connect
	} else {
		cfg |= nrf.P0_PIN_CNF_INPUT_Disconnect
	}
	nrf.P0.PIN_CNF[p.Pin] = nrf.RegValue(cfg)
}

func (p GPIO) Set(high bool) {
	if high {
		nrf.P0.OUTSET = 1 << p.Pin
	} else {
		nrf.P0.OUTCLR = 1 << p.Pin
	}
}

// Get returns the current value of a GPIO pin.
func (p GPIO) Get() (value bool) {
	// TODO: implement
	return
}
