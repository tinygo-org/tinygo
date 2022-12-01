//go:build arduino_leonardo
// +build arduino_leonardo

package machine

// Return the current CPU frequency in hertz.
func CPUFrequency() uint32 {
	return 16000000
}

// Digital pins, marked as plain numbers on the board.
const (
	D0  = PD2 // RX
	D1  = PD3 // TX
	D2  = PD1
	D3  = PD0
	D4  = PD4
	D5  = PC6
	D6  = PD7
	D7  = PE6
	D8  = PB4
	D9  = PB5
	D10 = PB6
	D11 = PB7
	D12 = PD6
	D13 = PC7
)

// LED on the Arduino
const LED Pin = D13

// ADC on the Arduino
const (
	ADC0 Pin = PF7
	ADC1 Pin = PF6
	ADC2 Pin = PF5
	ADC3 Pin = PF4
	ADC4 Pin = PF1
	ADC5 Pin = PF0
)
