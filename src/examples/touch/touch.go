//go:build esp32 || esp32s3_supermini || feather_esp32s3_tft || xiao_esp32s3

package main

import (
	"machine"
	"machine/touch"
	"time"
)

func main() {
	led := machine.LED
	led.Configure(machine.PinConfig{Mode: machine.PinOutput})

	pad := touch.Pad{Pin: machine.GPIO4}
	if err := pad.Configure(); err != nil {
		println(err.Error())
		return
	}

	// Measure the untouched value first, so do not touch the pad at startup.
	time.Sleep(100 * time.Millisecond)
	var sum uint32
	for i := 0; i < 16; i++ {
		sum += pad.Get()
		time.Sleep(10 * time.Millisecond)
	}
	base := sum / 16

	// The count drops on the ESP32 and rises on the ESP32-S3 when touched.
	for {
		val := pad.Get()
		diff := val - base
		if val < base {
			diff = base - val
		}
		touched := val != 0 && diff > base/5
		led.Set(touched)
		println(val, touched)
		time.Sleep(100 * time.Millisecond)
	}
}
