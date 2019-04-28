// +build !avr,!nrf,!sam,!stm32

package machine

// Dummy machine package, filled with no-ops.

type PinMode uint8

const (
	PinInput PinMode = iota
	PinOutput
)

// Fake LED numbers, for testing.
const (
	LED  Pin = LED1
	LED1 Pin = 0
	LED2 Pin = 0
	LED3 Pin = 0
	LED4 Pin = 0
)

// Fake button numbers, for testing.
const (
	BUTTON  Pin = BUTTON1
	BUTTON1 Pin = 0
	BUTTON2 Pin = 0
	BUTTON3 Pin = 0
	BUTTON4 Pin = 0
)

func (p Pin) Configure(config PinConfig) {
}

func (p Pin) Set(value bool) {
}

func (p Pin) Get() bool {
	return false
}
