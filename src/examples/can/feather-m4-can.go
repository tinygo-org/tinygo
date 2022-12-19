//go:build feather_m4_can

package main

import (
	"machine"
)

func init() {
	// power on the CAN Transceiver
	// https://learn.adafruit.com/adafruit-feather-m4-can-express/pinouts#can-bus-3078990-8
	boost_en := machine.BOOST_EN
	boost_en.Configure(machine.PinConfig{Mode: machine.PinOutput})
	boost_en.High()
}
