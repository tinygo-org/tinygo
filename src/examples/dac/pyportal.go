//go:build pyportal

package main

import (
	"machine"
)

func init() {
	enable := machine.SPK_SD
	enable.Configure(machine.PinConfig{Mode: machine.PinOutput})
	enable.Set(true)
}
