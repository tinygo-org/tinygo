// +build !avr,!nrf,!sam,!stm32

package machine

// Dummy machine package, filled with no-ops.

type GPIOMode uint8

const (
	GPIO_INPUT = iota
	GPIO_OUTPUT
)

// Fake LED numbers, for testing.
const (
	LED  = LED1
	LED1 = 0
	LED2 = 0
	LED3 = 0
	LED4 = 0
)

// Fake button numbers, for testing.
const (
	BUTTON  = BUTTON1
	BUTTON1 = 0
	BUTTON2 = 0
	BUTTON3 = 0
	BUTTON4 = 0
)

func (p GPIO) Configure(config GPIOConfig) {
}

func (p GPIO) Set(value bool) {
}

func (p GPIO) Get() bool {
	return false
}
