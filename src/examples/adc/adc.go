package main

import (
	"fmt"
	"machine"
	"time"
)

// ADC с пина 4 (GPIO4 = ADC1 ch3). raw 0..4095, V = raw/4095*3.3

func main() {
	time.Sleep(time.Second * 1)

	sensor := machine.ADC{machine.GPIO4}
	sensor.Configure(machine.ADCConfig{})

	println("ADC from GPIO4 (pin 4)...")

	for {
		val := sensor.Get()
		raw12 := val >> 4
		v := float64(raw12) / 4095.0 * 3.3
		fmt.Printf("raw=%d  V~%.3f\n", raw12, v)
		time.Sleep(time.Millisecond * 100)
	}
}
