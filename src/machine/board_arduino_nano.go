//go:build arduino_nano
// +build arduino_nano

package machine

// Return the current CPU frequency in hertz.
func CPUFrequency() uint32 {
	return 16000000
}

// Digital pins.
const (
	D0  = PD0 // RX0
	D1  = PD1 // TX1
	D2  = PD2
	D3  = PD3
	D4  = PD4
	D5  = PD5
	D6  = PD6
	D7  = PD7
	D8  = PB0
	D9  = PB1
	D10 = PB2
	D11 = PB3
	D12 = PB4
	D13 = PB5
)

// LED on the Arduino
const LED Pin = D13

// ADC on the Arduino
const (
	ADC0 Pin = PC0
	ADC1 Pin = PC1
	ADC2 Pin = PC2
	ADC3 Pin = PC3
	ADC4 Pin = PC4 // Used by TWI for SDA
	ADC5 Pin = PC5 // Used by TWI for SCL
)

// UART pins
const (
	UART_TX_PIN Pin = PD1
	UART_RX_PIN Pin = PD0
)
