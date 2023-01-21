//go:build rp2040

package main

// This example shows how to send a board to a low power mode (deep sleep) and wake it up on timer.
//
// There are 3 states the board can be in this example:
// - Base power: on-board LED is off, does not consume power (5 seconds);
// - High power: on-board LED shines, consumes power (5 seconds);
// -  Low power: almost everything is off but crystal oscilator, rtc, sys and ref clocks (5 seconds).
//
// Measured power consumption for Arduino Nano RP2040 Connect board powered from 2S battery (~8.2v)
// - Base power: 13.55 mA (0.11W)  bare minimum on fully powered up RP2040 chip
// - High power: 15.55 mA (0.13W)  on-board LED consumes 2 mA (same for power LED, see below)
// - Deep sleep:  4.00 mA (0.03W)  power LED consumes 2 mA
//                2.00 mA (0.015W) power LED removed => ~month life time, see below
//
// Life expectancy on 2S 2200 Li-Ion battery.
// 4.2v->3.2v, avg 3.7v per cell (7.4v total).
// Remaining charge ~1000mAh not to ruin the battery => 1.2Ah consumable.
//
// Expected number of hours = 3.7*2*1.2 / w, there "w" is power consumption, in Watts.
//
// High power:  68h (~3 days)
// Base power:  81h (~3.5 days)
// Deep sleep: 296h (~12 days)
//             592h (~24 days)

import (
	"machine"
	"time"
)

const led = machine.LED

func main() {

	led.Configure(machine.PinConfig{Mode: machine.PinOutput})

	// Pause and let user connect to serial console
	time.Sleep(5 * time.Second)

	// Fire RTC interrupt every 15 seconds: ~10 seconds we spend in high+base power and sleep the rest
	machine.RTC.SetInterrupt(10+5, true, nil)

	for {

		// Base power
		// Nano RP2040 Connect: 0.11W
		log("Base power, 5 sec, 0.11W")
		led.Low()
		time.Sleep(5 * time.Second) // light sleep

		// High power
		// Nano RP2040 Connect: 0.13W
		log("High power, 5 sec, 0.13W")
		led.High()
		time.Sleep(5 * time.Second) // light sleep

		// Low power
		// Nano RP2040 Connect: 0.03W
		log(" Low power, 5 sec, 0.03W")
		led.Low()
		// machine.RTC.SetInterrupt(10, false, nil) // alternatively non-recurring RTC interrupt can be scheduled every time
		machine.Sleep() // deep sleep

	}

}

func log(message string) {
	time.Sleep(10 * time.Millisecond) // ensure output subsystem fully initialised
	println(time.Now().Format(time.RFC3339) + " " + message)
	time.Sleep(10 * time.Millisecond) // ensure output subsystem doe not go sleep before message printed out
}
