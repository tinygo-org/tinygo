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
