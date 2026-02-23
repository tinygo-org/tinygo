package main

import (
	"fmt"
	"machine"
	"time"
)

// Reads ADC, prints raw (0..4095) and V like Arduino. ESP32-S3: ADC1 = GPIO1..10, ADC2 = GPIO11..20.
// На многих платах GPIO2 занят (подтяжка/загрузка) — если масса не даёт 0, пробуй GPIO4 или другой свободный пин.

func main() {
	time.Sleep(time.Second * 1)

	sensor := machine.ADC{machine.GPIO2}
	sensor.Configure(machine.ADCConfig{})

	println("ADC read from GPIO2...")

	for {
		val := sensor.Get()
		raw12 := val >> 4
		v := float64(raw12) / 4095.0 * 3.3
		fmt.Printf("raw=%d  V~%.3f\n", raw12, v)
		time.Sleep(time.Millisecond * 100)
	}
}
