/*
Read 3 rp2040 ADC channels
   ADC0_CH     - GPIO26 Raw reading every 100ms
   ADC1_CH     - GPIO27 Voltage reading every 3s
   ADC_TEMP_CH - Temperature sensor reading every serial output (I.e. every second)
*/
package main

import (
	"fmt"
	"machine"
	"time"
)

type celsius float32

func (c celsius) String() string {
	return fmt.Sprintf("%4.1f℃", c)
}

type volts float32

func (v volts) String() string {
	return fmt.Sprintf("%2.1fV", v)
}

var (
	adc0Raw   uint16
	adc1Volts volts
)

func readRawOften(c machine.ADCChannel) {
	for {
		adc0Raw = c.GetOnce()
		time.Sleep(100 * time.Millisecond)
	}
}

func readVoltsInfrequently(c machine.ADCChannel) {
	for {
		adc1Volts = volts(c.GetVoltage())
		time.Sleep(3000 * time.Millisecond)
	}
}

func main() {
	machine.InitADC()
	a0 := machine.ADC0_CH    // GPIO26 input
	a1 := machine.ADC1_CH    // GPIO27 input
	t := machine.ADC_TEMP_CH // Internal Temperature sensor
	// Configure sets the GPIOs to PinAnalog mode
	a0.Configure(machine.ADCChConfig{})
	a1.Configure(machine.ADCChConfig{})
	// Configure powers on the temperature sensor
	t.Configure(machine.ADCChConfig{})

	// Safe to read concurrently
	go readRawOften(a0)
	go readVoltsInfrequently(a1)

	for {
		fmt.Printf("ADC0: %5d ADC1: %v Temp: %v\n\r", adc0Raw, adc1Volts, celsius(t.GetTemp()))
		time.Sleep(1000 * time.Millisecond)
	}
}
