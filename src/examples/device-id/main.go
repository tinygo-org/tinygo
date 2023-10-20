package main

import (
	"encoding/hex"
	"machine"
	"time"
)

func main() {
	time.Sleep(2 * time.Second)

	// For efficiency, it's best to get the device ID once and cache it
	// (e.g. on RP2040 XIP flash and interrupts disabled for period of
	// retrieving the hardware ID from ROM chip)
	id := machine.DeviceID()

	for {
		println("Device ID:", hex.EncodeToString(id))
		time.Sleep(1 * time.Second)
	}
}
