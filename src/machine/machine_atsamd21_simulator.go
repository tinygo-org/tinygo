//go:build !baremetal && (gemma_m0 || qtpy || trinket_m0 || arduino_mkr1000 || arduino_mkrwifi1010 || arduino_nano33 || arduino_zero || circuitplay_express || feather_m0_express || feather_m0 || itsybitsy_m0 || p1am_100 || xiao)

// Simulated atsamd21 chips.

package machine

// The timer channels/pins match the hardware, and encode the same information
// as pinTimerMapping but in a more generic (less efficient) way.

var TCC0 = &timerType{
	instance:   0,
	frequency:  48e6,
	bits:       24,
	prescalers: []int{1, 2, 4, 8, 16, 64, 256, 1024},
	channelPins: [][]Pin{
		{PA04, PA08, PB10, PA14, PB16, PA22, PB30}, // channel 0
		{PA05, PA09, PB11, PA15, PB17, PA23, PB31}, // channel 1
		{PA10, PB12, PA12, PA16, PA18, PA20},       // channel 2
		{PA11, PB13, PA13, PA17, PA19, PA21},       // channel 3
	},
}

var TCC1 = &timerType{
	instance:   1,
	frequency:  48e6,
	bits:       24,
	prescalers: []int{1, 2, 4, 8, 16, 64, 256, 1024},
	channelPins: [][]Pin{
		{PA06, PA10, PA30}, // channel 0
		{PA07, PA11, PA31}, // channel 1
		{PA08, PA24, PB30}, // channel 2
		{PA09, PA25, PB31}, // channel 3
	},
}

var TCC2 = &timerType{
	instance:   2,
	frequency:  48e6,
	bits:       16,
	prescalers: []int{1, 2, 4, 8, 16, 64, 256, 1024},
	channelPins: [][]Pin{
		{PA00, PA12, PA16}, // channel 0
		{PA01, PA13, PA17}, // channel 1
	},
}

var (
	// According to the datasheet, only some pins have I2C support. However it
	// looks like many boards just use any SERCOM I2C instance, even if the
	// datasheet says those don't support I2C. I guess they do work in practice,
	// then.
	// These are:
	//   * PA00/PA01 for the Adafruit Circuit Playground Express (I2C1, SERCOM1).
	//   * PB02/PB03 for the Adafruit Circuit Playground Express (I2C0, SERCOM5).
	//   * PB08/PB09 for the Arduino Nano 33 IoT (I2C0, SERCOM4).
	// https://cdn.sparkfun.com/datasheets/Dev/Arduino/Boards/Atmel-42181-SAM-D21_Datasheet.pdf
	sercomI2CM0 = &I2C{Bus: 0, PinsSDA: []Pin{PA08}, PinsSCL: []Pin{PA09}}
	sercomI2CM1 = &I2C{Bus: 1, PinsSDA: []Pin{PA00, PA16}, PinsSCL: []Pin{PA01, PA17}}
	sercomI2CM2 = &I2C{Bus: 2, PinsSDA: []Pin{PA08, PA12}, PinsSCL: []Pin{PA09, PA13}}
	sercomI2CM3 = &I2C{Bus: 3, PinsSDA: []Pin{PA16, PA22}, PinsSCL: []Pin{PA17, PA23}}
	sercomI2CM4 = &I2C{Bus: 4, PinsSDA: []Pin{PA12, PB08, PB12}, PinsSCL: []Pin{PA13, PB09, PB13}}
	sercomI2CM5 = &I2C{Bus: 5, PinsSDA: []Pin{PA22, PB02, PB16, PB30}, PinsSCL: []Pin{PA23, PB03, PB17, PB31}}
)
