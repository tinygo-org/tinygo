//go:build circuitplay_express
// +build circuitplay_express

package main

import (
	"machine"
)

func init() {
	enable := machine.PA30
	enable.Configure(machine.PinConfig{Mode: machine.PinOutput})
	enable.Set(true)
}
