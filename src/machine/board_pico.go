// +build pico

package machine

// GPIO pins
const (
	GP0  Pin = 0
	GP1  Pin = 1
	GP2  Pin = 2
	GP3  Pin = 3
	GP4  Pin = 4
	GP5  Pin = 5
	GP6  Pin = 6
	GP7  Pin = 7
	GP8  Pin = 8
	GP9  Pin = 9
	GP10 Pin = 10
	GP11 Pin = 11
	GP12 Pin = 12
	GP13 Pin = 13
	GP14 Pin = 14
	GP15 Pin = 15
	GP16 Pin = 16
	GP17 Pin = 17
	GP18 Pin = 18
	GP19 Pin = 19
	GP20 Pin = 20
	GP21 Pin = 21
	GP22 Pin = 22
	GP23 Pin = 23
	GP24 Pin = 24
	GP25 Pin = 25
	GP26 Pin = 26
	GP27 Pin = 27
	GP28 Pin = 28
	GP29 Pin = 29

	// Onboard LED
	LED Pin = GP25

	// Analog pins
	ADC0 = GP26
	ADC1 = GP27
	ADC2 = GP28
	ADC3 = GP29

	// Onboard crystal oscillator frequency, in MHz.
	xoscFreq = 12 // MHz
)
