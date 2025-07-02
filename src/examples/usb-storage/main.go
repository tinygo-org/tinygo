package main

import (
	"machine"
	"machine/usb/msc"
	"time"
)

func main() {
	msc.Port(machine.Flash)

	for {
		time.Sleep(2 * time.Second)
	}
}
