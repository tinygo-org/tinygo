//go:build rp2350

package machine

import (
	"device/rp"
)

const deviceName = rp.Device

//go:linkname machineInit runtime.machineInit
func machineInit() {
}

// GPIO pins
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
	GPIO22 Pin = 22
	GPIO23 Pin = 23
	GPIO24 Pin = 24
	GPIO25 Pin = 25
	GPIO26 Pin = 26
	GPIO27 Pin = 27
	GPIO28 Pin = 28
	GPIO29 Pin = 29

	LED = GPIO26
)

const (
	PinOutput PinMode = iota
)

// PinMode sets the direction and pull mode of the pin. For example, PinOutput
// sets the pin as an output and PinInputPullup sets the pin as an input with a
// pull-up.
// type PinMode uint8

// type PinConfig struct {
// 	Mode PinMode
// }

func (p Pin) Set(b bool)          {}
func (p Pin) Get() bool           { return false }
func (p Pin) Configure(PinConfig) {}
func putchar(b byte)              {}

type timeUnit int64

// ticks returns the number of ticks (microseconds) elapsed since power up.
//
//go:linkname ticks runtime.machineTicks
func ticks() uint64 {
	return 0
}

//go:linkname machineLightSleep runtime.machineLightSleep
func machineLightSleep(uint64) {}

func nanosecondsToTicks(ns int64) timeUnit {
	return timeUnit(ns / 1000)
}

func ticksToNanoseconds(ticks timeUnit) int64 {
	return int64(ticks) * 1000
}
