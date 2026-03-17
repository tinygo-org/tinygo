//go:build !baremetal && (xiao_esp32c3 || xiao_esp32s3)

// Simulator support for ESP32-C3 and ESP32-S3 based boards.

package machine

// Hardware pin numbers
const (
	GPIO0  Pin = 0
	GPIO1  Pin = 1
	GPIO2  Pin = 2
	GPIO3  Pin = 3
	GPIO4  Pin = 4
	GPIO5  Pin = 5
	GPIO6  Pin = 6
	GPIO7  Pin = 7
	GPIO8  Pin = 8
	GPIO9  Pin = 9
	GPIO10 Pin = 10
	GPIO11 Pin = 11
	GPIO12 Pin = 12
	GPIO13 Pin = 13
	GPIO14 Pin = 14
	GPIO15 Pin = 15
	GPIO16 Pin = 16
	GPIO17 Pin = 17
	GPIO18 Pin = 18
	GPIO19 Pin = 19
	GPIO20 Pin = 20
	GPIO21 Pin = 21
	GPIO26 Pin = 26
	GPIO27 Pin = 27
	GPIO28 Pin = 28
	GPIO29 Pin = 29
	GPIO30 Pin = 30
	GPIO31 Pin = 31
	GPIO32 Pin = 32
	GPIO33 Pin = 33
	GPIO34 Pin = 34
	GPIO35 Pin = 35
	GPIO36 Pin = 36
	GPIO37 Pin = 37
	GPIO38 Pin = 38
	GPIO39 Pin = 39
	GPIO40 Pin = 40
	GPIO41 Pin = 41
	GPIO42 Pin = 42
	GPIO43 Pin = 43
	GPIO44 Pin = 44
	GPIO45 Pin = 45
	GPIO46 Pin = 46
	GPIO47 Pin = 47
	GPIO48 Pin = 48
)

const (
	ADC0  Pin = GPIO1
	ADC2  Pin = GPIO2
	ADC3  Pin = GPIO3
	ADC4  Pin = GPIO4
	ADC5  Pin = GPIO5
	ADC6  Pin = GPIO6
	ADC7  Pin = GPIO7
	ADC8  Pin = GPIO8
	ADC9  Pin = GPIO9
	ADC10 Pin = GPIO10
	ADC11 Pin = GPIO11
	ADC12 Pin = GPIO12
	ADC13 Pin = GPIO13
	ADC14 Pin = GPIO14
	ADC15 Pin = GPIO15
	ADC16 Pin = GPIO16
	ADC17 Pin = GPIO17
	ADC18 Pin = GPIO18
	ADC19 Pin = GPIO19
	ADC20 Pin = GPIO20
)

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

var I2C0 = &I2C{Bus: 0}
