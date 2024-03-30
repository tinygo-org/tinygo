//go:build rp2040 || ae_rp2040 || badger2040 || challenger_rp2040 || feather_rp2040 || gopher_badge || kb2040 || macropad_rp2040 || nano_rp2040 || pico || qtpy_rp2040 || thingplus_rp2040 || thumby || tufty2040 || waveshare_rp2040_zero || xiao_rp2040

package machine

const (
	// GPIO pins
	GPIO0  Pin = 0  // peripherals: PWM0 channel A
	GPIO1  Pin = 1  // peripherals: PWM0 channel B
	GPIO2  Pin = 2  // peripherals: PWM1 channel A
	GPIO3  Pin = 3  // peripherals: PWM1 channel B
	GPIO4  Pin = 4  // peripherals: PWM2 channel A
	GPIO5  Pin = 5  // peripherals: PWM2 channel B
	GPIO6  Pin = 6  // peripherals: PWM3 channel A
	GPIO7  Pin = 7  // peripherals: PWM3 channel B
	GPIO8  Pin = 8  // peripherals: PWM4 channel A
	GPIO9  Pin = 9  // peripherals: PWM4 channel B
	GPIO10 Pin = 10 // peripherals: PWM5 channel A
	GPIO11 Pin = 11 // peripherals: PWM5 channel B
	GPIO12 Pin = 12 // peripherals: PWM6 channel A
	GPIO13 Pin = 13 // peripherals: PWM6 channel B
	GPIO14 Pin = 14 // peripherals: PWM7 channel A
	GPIO15 Pin = 15 // peripherals: PWM7 channel B
	GPIO16 Pin = 16 // peripherals: PWM0 channel A
	GPIO17 Pin = 17 // peripherals: PWM0 channel B
	GPIO18 Pin = 18 // peripherals: PWM1 channel A
	GPIO19 Pin = 19 // peripherals: PWM1 channel B
	GPIO20 Pin = 20 // peripherals: PWM2 channel A
	GPIO21 Pin = 21 // peripherals: PWM2 channel B
	GPIO22 Pin = 22 // peripherals: PWM3 channel A
	GPIO23 Pin = 23 // peripherals: PWM3 channel B
	GPIO24 Pin = 24 // peripherals: PWM4 channel A
	GPIO25 Pin = 25 // peripherals: PWM4 channel B
	GPIO26 Pin = 26 // peripherals: PWM5 channel A
	GPIO27 Pin = 27 // peripherals: PWM5 channel B
	GPIO28 Pin = 28 // peripherals: PWM6 channel A
	GPIO29 Pin = 29 // peripherals: PWM6 channel B

	// Analog pins
	ADC0 Pin = GPIO26
	ADC1 Pin = GPIO27
	ADC2 Pin = GPIO28
	ADC3 Pin = GPIO29
)
