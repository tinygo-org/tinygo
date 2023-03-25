//go:build feather_nrf52840

package main

import "machine"

const (
	TARGET_SCL = machine.A5
	TARGET_SDA = machine.A4
)

var (
	controller = machine.I2C0
	target     = machine.I2C1
)
