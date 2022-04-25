// Reads multiple rp2040 ADC channels concurrently. Including the internal temperature sensor

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

// rp2040 ADC is 12 bits. Reading are shifted <<4 to fill the 16-bit range.
var adcReading [3]uint16

func readADC(a machine.ADC, w time.Duration, i int) {
	for {
		adcReading[i] = a.Get()
		time.Sleep(w)
	}
}

func main() {
	machine.InitADC()
	a0 := machine.ADC{machine.ADC0} // GPIO26 input
	a1 := machine.ADC{machine.ADC1} // GPIO27 input
	a2 := machine.ADC{machine.ADC2} // GPIO28 input
	t := machine.ADC_TEMP_SENSOR    // Internal Temperature sensor
	// Configure sets the GPIOs to PinAnalog mode
	a0.Configure(machine.ADCConfig{})
	a1.Configure(machine.ADCConfig{})
	a2.Configure(machine.ADCConfig{})
	// Configure powers on the temperature sensor
	t.Configure(machine.ADCConfig{})

	// Safe to read concurrently
	go readADC(a0, 10*time.Millisecond, 0)
	go readADC(a1, 17*time.Millisecond, 1)
	go readADC(a2, 29*time.Millisecond, 2)

	for {
		fmt.Printf("ADC0: %5d ADC1: %5d ADC2: %5d Temp: %v\n\r", adcReading[0], adcReading[1], adcReading[2], celsius(float32(t.ReadTemperature())/1000))
		time.Sleep(1000 * time.Millisecond)
	}
}
