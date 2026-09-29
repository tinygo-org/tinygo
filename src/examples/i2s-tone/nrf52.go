//go:build nrf52 || nrf52833 || (nrf52840 && !itsybitsy_nrf52840)

package main

import "machine"

const (
	pinSCK = machine.P0_03
	pinWS  = machine.P0_04
	pinSDO = machine.P0_28
)
