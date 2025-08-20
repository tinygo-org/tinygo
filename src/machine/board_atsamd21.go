//go:build (sam && atsamd21) || arduino_nano33 || circuitplay_express

package machine

// Return the current CPU frequency in hertz.
func CPUFrequency() uint32 {
	return 48000000
}

// Note that the below pins have a few I2C pins listed that according to the
// datasheet don't support I2C yet they are used in practice on boards from
// Adafruit and Arduino. See machine_atsamd21_simulator.go for details.

// Hardware pins
const (
	PA00 Pin = 0 // peripherals: TCC2 channel 0, sercomI2CM1 SDA
	PA01 Pin = 1 // peripherals: TCC2 channel 1, sercomI2CM1 SCL
	PA02 Pin = 2
	PA03 Pin = 3
	PA04 Pin = 4  // peripherals: TCC0 channel 0
	PA05 Pin = 5  // peripherals: TCC0 channel 1
	PA06 Pin = 6  // peripherals: TCC1 channel 0
	PA07 Pin = 7  // peripherals: TCC1 channel 1
	PA08 Pin = 8  // peripherals: TCC0 channel 0, TCC1 channel 2, sercomI2CM0 SDA, sercomI2CM2 SDA
	PA09 Pin = 9  // peripherals: TCC0 channel 1, TCC1 channel 3, sercomI2CM0 SCL, sercomI2CM2 SCL
	PA10 Pin = 10 // peripherals: TCC1 channel 0, TCC0 channel 2
	PA11 Pin = 11 // peripherals: TCC1 channel 1, TCC0 channel 3
	PA12 Pin = 12 // peripherals: TCC2 channel 0, TCC0 channel 2, sercomI2CM2 SDA, sercomI2CM4 SDA
	PA13 Pin = 13 // peripherals: TCC2 channel 1, TCC0 channel 3, sercomI2CM2 SCL, sercomI2CM4 SCL
	PA14 Pin = 14 // peripherals: TCC0 channel 0
	PA15 Pin = 15 // peripherals: TCC0 channel 1
	PA16 Pin = 16 // peripherals: TCC2 channel 0, TCC0 channel 2, sercomI2CM1 SDA, sercomI2CM3 SDA
	PA17 Pin = 17 // peripherals: TCC2 channel 1, TCC0 channel 3, sercomI2CM1 SCL, sercomI2CM3 SCL
	PA18 Pin = 18 // peripherals: TCC0 channel 2
	PA19 Pin = 19 // peripherals: TCC0 channel 3
	PA20 Pin = 20 // peripherals: TCC0 channel 2
	PA21 Pin = 21 // peripherals: TCC0 channel 3
	PA22 Pin = 22 // peripherals: TCC0 channel 0, sercomI2CM3 SDA, sercomI2CM5 SDA
	PA23 Pin = 23 // peripherals: TCC0 channel 1, sercomI2CM3 SCL, sercomI2CM5 SCL
	PA24 Pin = 24 // peripherals: TCC1 channel 2
	PA25 Pin = 25 // peripherals: TCC1 channel 3
	PA26 Pin = 26
	PA27 Pin = 27
	PA28 Pin = 28
	PA29 Pin = 29
	PA30 Pin = 30 // peripherals: TCC1 channel 0
	PA31 Pin = 31 // peripherals: TCC1 channel 1
	PB00 Pin = 32
	PB01 Pin = 33
	PB02 Pin = 34 // peripherals: sercomI2CM5 SDA
	PB03 Pin = 35 // peripherals: sercomI2CM5 SCL
	PB04 Pin = 36
	PB05 Pin = 37
	PB06 Pin = 38
	PB07 Pin = 39
	PB08 Pin = 40 // peripherals: sercomI2CM4 SDA
	PB09 Pin = 41 // peripherals: sercomI2CM4 SCL
	PB10 Pin = 42 // peripherals: TCC0 channel 0
	PB11 Pin = 43 // peripherals: TCC0 channel 1
	PB12 Pin = 44 // peripherals: TCC0 channel 2, sercomI2CM4 SDA
	PB13 Pin = 45 // peripherals: TCC0 channel 3, sercomI2CM4 SCL
	PB14 Pin = 46
	PB15 Pin = 47
	PB16 Pin = 48 // peripherals: TCC0 channel 0, sercomI2CM5 SDA
	PB17 Pin = 49 // peripherals: TCC0 channel 1, sercomI2CM5 SCL
	PB18 Pin = 50
	PB19 Pin = 51
	PB20 Pin = 52
	PB21 Pin = 53
	PB22 Pin = 54
	PB23 Pin = 55
	PB24 Pin = 56
	PB25 Pin = 57
	PB26 Pin = 58
	PB27 Pin = 59
	PB28 Pin = 60
	PB29 Pin = 61
	PB30 Pin = 62 // peripherals: TCC0 channel 0, TCC1 channel 2, sercomI2CM5 SDA
	PB31 Pin = 63 // peripherals: TCC0 channel 1, TCC1 channel 3, sercomI2CM5 SCL
)
