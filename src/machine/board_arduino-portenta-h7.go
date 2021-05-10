// +build portenta_h7

// Peripheral and pin mappings for Arduino Portenta H7:
//   https://store.arduino.cc/usa/portenta-h7
//
// This only defines peripherals/pins available on the MKR header. For access
// to pins on the 80-pin high-density connectors, you must use the CPU port pin
// identifiers (e.g., PH15) defined in machine_stm32.go.

package machine

// CPU identifies the current operating core of the dual-core MCU.
// The Cortex-M7 is core 0, and the Cortex-M4 is core 1.
const CPU = cpuCore

// GPIO Pins
const (
	D0  = PH15
	D1  = PK01
	D2  = PJ11
	D3  = PG07
	D4  = PC07
	D5  = PC06
	D6  = PA08
	D7  = PI00
	D8  = PC03
	D9  = PI01
	D10 = PC02
	D11 = PH08
	D12 = PH07
	D13 = PA10
	D14 = PA09
)

// Analog pins
const (
	A0 = PA00 // C
	A1 = PA01 // C
	A2 = PC02 // C
	A3 = PC03 // C
	A4 = PC02 // ALT0
	A5 = PC03 // ALT0
	A6 = PA04
	A7 = PA06
)

// On-board RGB LED pins
const (
	LED  = LEDR
	LEDR = PK05
	LEDG = PK06
	LEDB = PK07
)
