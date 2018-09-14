// +build nrf

package machine

import (
	"device/nrf"
)

type GPIOMode uint8

const (
	GPIO_INPUT          = (nrf.P0_PIN_CNF_DIR_Input << nrf.P0_PIN_CNF_DIR_Pos) | (nrf.P0_PIN_CNF_INPUT_Connect << nrf.P0_PIN_CNF_INPUT_Pos)
	GPIO_INPUT_PULLUP   = GPIO_INPUT | (nrf.P0_PIN_CNF_PULL_Pullup << nrf.P0_PIN_CNF_PULL_Pos)
	GPIO_INPUT_PULLDOWN = GPIO_INPUT | (nrf.P0_PIN_CNF_PULL_Pulldown << nrf.P0_PIN_CNF_PULL_Pos)
	GPIO_OUTPUT         = (nrf.P0_PIN_CNF_DIR_Output << nrf.P0_PIN_CNF_DIR_Pos) | (nrf.P0_PIN_CNF_INPUT_Disconnect << nrf.P0_PIN_CNF_INPUT_Pos)
)

// LEDs on the PCA10040 (nRF52832 dev board)
const (
	LED  = LED1
	LED1 = 17
	LED2 = 18
	LED3 = 19
	LED4 = 20
)

// Buttons on the PCA10040 (nRF52832 dev board)
const (
	BUTTON  = BUTTON1
	BUTTON1 = 13
	BUTTON2 = 14
	BUTTON3 = 15
	BUTTON4 = 16
)

// Configure this pin with the given configuration.
func (p GPIO) Configure(config GPIOConfig) {
	cfg := config.Mode | nrf.P0_PIN_CNF_DRIVE_S0S1 | nrf.P0_PIN_CNF_SENSE_Disabled
	nrf.P0.PIN_CNF[p.Pin] = nrf.RegValue(cfg)
}

// Set the pin to high or low.
// Warning: only use this on an output pin!
func (p GPIO) Set(high bool) {
	if high {
		nrf.P0.OUTSET = 1 << p.Pin
	} else {
		nrf.P0.OUTCLR = 1 << p.Pin
	}
}

// Get returns the current value of a GPIO pin.
func (p GPIO) Get() bool {
	return (nrf.P0.IN>>p.Pin)&1 != 0
}
