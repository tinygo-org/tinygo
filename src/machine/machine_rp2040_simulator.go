//go:build !baremetal && (ae_rp2040 || badger2040_w || badger2040 || challenger_rp2040 || elecrow_rp2040 || feather_rp2040 || gopher_badge || kb2040 || macropad_rp2040 || nano_rp2040 || pico || qtpy_rp2040 || thingplus_rp2040 || thumby || trinkey_qt2040 || tufty2040 || waveshare_rp2040_tiny || waveshare_rp2040_zero || xiao_rp2040)

// Simulator support for the RP2040.
//
// This is *only* for the RP2040. RP2350 is a different chip with slightly
// different characteristics.

package machine

var PWM0 = &timerType{
	instance:   0,
	frequency:  200e6,
	bits:       16,
	prescalers: []int{1, 2, 4, 8, 16, 32, 64, 128, 256}, // actually a continuing range, TODO
	channelPins: [][]Pin{
		{GPIO0, GPIO16}, // channel A (0)
		{GPIO1, GPIO17}, // channel B (1)
	},
}

var PWM1 = &timerType{
	instance:   0,
	frequency:  200e6,
	bits:       16,
	prescalers: []int{1, 2, 4, 8, 16, 32, 64, 128, 256},
	channelPins: [][]Pin{
		{GPIO2, GPIO18}, // channel A (0)
		{GPIO3, GPIO19}, // channel B (1)
	},
}

var PWM2 = &timerType{
	instance:   0,
	frequency:  200e6,
	bits:       16,
	prescalers: []int{1, 2, 4, 8, 16, 32, 64, 128, 256},
	channelPins: [][]Pin{
		{GPIO4, GPIO20}, // channel A (0)
		{GPIO5, GPIO21}, // channel B (1)
	},
}

var PWM3 = &timerType{
	instance:   0,
	frequency:  200e6,
	bits:       16,
	prescalers: []int{1, 2, 4, 8, 16, 32, 64, 128, 256},
	channelPins: [][]Pin{
		{GPIO6, GPIO22}, // channel A (0)
		{GPIO7, GPIO23}, // channel B (1)
	},
}

var PWM4 = &timerType{
	instance:   0,
	frequency:  200e6,
	bits:       16,
	prescalers: []int{1, 2, 4, 8, 16, 32, 64, 128, 256},
	channelPins: [][]Pin{
		{GPIO8, GPIO24}, // channel A (0)
		{GPIO9, GPIO25}, // channel B (1)
	},
}

var PWM5 = &timerType{
	instance:   0,
	frequency:  200e6,
	bits:       16,
	prescalers: []int{1, 2, 4, 8, 16, 32, 64, 128, 256},
	channelPins: [][]Pin{
		{GPIO10, GPIO26}, // channel A (0)
		{GPIO11, GPIO27}, // channel B (1)
	},
}

var PWM6 = &timerType{
	instance:   0,
	frequency:  200e6,
	bits:       16,
	prescalers: []int{1, 2, 4, 8, 16, 32, 64, 128, 256},
	channelPins: [][]Pin{
		{GPIO12, GPIO28}, // channel A (0)
		{GPIO13, GPIO29}, // channel B (1)
	},
}

var PWM7 = &timerType{
	instance:   0,
	frequency:  200e6,
	bits:       16,
	prescalers: []int{1, 2, 4, 8, 16, 32, 64, 128, 256},
	channelPins: [][]Pin{
		{GPIO14}, // channel A (0)
		{GPIO15}, // channel B (1)
	},
}

var I2C0 = &I2C{
	Bus:     0,
	PinsSCL: []Pin{GPIO1, GPIO5, GPIO9, GPIO13, GPIO17, GPIO21},
	PinsSDA: []Pin{GPIO0, GPIO4, GPIO8, GPIO12, GPIO16, GPIO20},
}

var I2C1 = &I2C{
	Bus:     0,
	PinsSCL: []Pin{GPIO3, GPIO7, GPIO11, GPIO15, GPIO19, GPIO27},
	PinsSDA: []Pin{GPIO2, GPIO6, GPIO10, GPIO14, GPIO18, GPIO26},
}
