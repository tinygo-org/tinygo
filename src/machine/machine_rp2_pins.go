//go:build rp2040 || rp2350 || gopher_badge || pico

package machine

const (
	// GPIO pins
	GPIO0  Pin = 0  // peripherals: PWM0 channel A, I2C0 SDA
	GPIO1  Pin = 1  // peripherals: PWM0 channel B, I2C0 SCL
	GPIO2  Pin = 2  // peripherals: PWM1 channel A, I2C1 SDA
	GPIO3  Pin = 3  // peripherals: PWM1 channel B, I2C1 SCL
	GPIO4  Pin = 4  // peripherals: PWM2 channel A, I2C0 SDA
	GPIO5  Pin = 5  // peripherals: PWM2 channel B, I2C0 SCL
	GPIO6  Pin = 6  // peripherals: PWM3 channel A, I2C1 SDA
	GPIO7  Pin = 7  // peripherals: PWM3 channel B, I2C1 SCL
	GPIO8  Pin = 8  // peripherals: PWM4 channel A, I2C0 SDA
	GPIO9  Pin = 9  // peripherals: PWM4 channel B, I2C0 SCL
	GPIO10 Pin = 10 // peripherals: PWM5 channel A, I2C1 SDA
	GPIO11 Pin = 11 // peripherals: PWM5 channel B, I2C1 SCL
	GPIO12 Pin = 12 // peripherals: PWM6 channel A, I2C0 SDA
	GPIO13 Pin = 13 // peripherals: PWM6 channel B, I2C0 SCL
	GPIO14 Pin = 14 // peripherals: PWM7 channel A, I2C1 SDA
	GPIO15 Pin = 15 // peripherals: PWM7 channel B, I2C1 SCL
	GPIO16 Pin = 16 // peripherals: PWM0 channel A, I2C0 SDA
	GPIO17 Pin = 17 // peripherals: PWM0 channel B, I2C0 SCL
	GPIO18 Pin = 18 // peripherals: PWM1 channel A, I2C1 SDA
	GPIO19 Pin = 19 // peripherals: PWM1 channel B, I2C1 SCL
	GPIO20 Pin = 20 // peripherals: PWM2 channel A, I2C0 SDA
	GPIO21 Pin = 21 // peripherals: PWM2 channel B, I2C0 SCL
	GPIO22 Pin = 22 // peripherals: PWM3 channel A
	GPIO23 Pin = 23 // peripherals: PWM3 channel B
	GPIO24 Pin = 24 // peripherals: PWM4 channel A
	GPIO25 Pin = 25 // peripherals: PWM4 channel B
	GPIO26 Pin = 26 // peripherals: PWM5 channel A, I2C1 SDA
	GPIO27 Pin = 27 // peripherals: PWM5 channel B, I2C1 SCL
	GPIO28 Pin = 28 // peripherals: PWM6 channel A
	GPIO29 Pin = 29 // peripherals: PWM6 channel B
)
