//go:build !baremetal && (bluemicro840 || circuitplay_bluefruit || clue_alpha || feather_nrf52840_sense || feather_nrf52840 || itsybitsy_nrf52840 || mdbt50qrx || nano_33_ble || nicenano || nrf52840_mdk || particle_3rd_gen || pca10056 || pca10059 || rak4631 || reelboard || xiao_ble)

// Simulator support for nrf52840 based boards.

package machine

// Channel values below are nil, so that they get filled in on the first use.
// This is the same as what happens on baremetal.

var PWM0 = &timerType{
	instance:   0,
	frequency:  16e6,
	bits:       15,
	prescalers: []int{1, 2, 4, 8, 16, 32, 64, 128},
	channelPins: [][]Pin{
		nil, // channel 0
		nil, // channel 1
		nil, // channel 2
		nil, // channel 3
	},
}

var PWM1 = &timerType{
	instance:   1,
	frequency:  16e6,
	bits:       15,
	prescalers: []int{1, 2, 4, 8, 16, 32, 64, 128},
	channelPins: [][]Pin{
		nil, // channel 0
		nil, // channel 1
		nil, // channel 2
		nil, // channel 3
	},
}

var PWM2 = &timerType{
	instance:   2,
	frequency:  16e6,
	bits:       15,
	prescalers: []int{1, 2, 4, 8, 16, 32, 64, 128},
	channelPins: [][]Pin{
		nil, // channel 0
		nil, // channel 1
		nil, // channel 2
		nil, // channel 3
	},
}

var PWM3 = &timerType{
	instance:   3,
	frequency:  16e6,
	bits:       15,
	prescalers: []int{1, 2, 4, 8, 16, 32, 64, 128},
	channelPins: [][]Pin{
		nil, // channel 0
		nil, // channel 1
		nil, // channel 2
		nil, // channel 3
	},
}
