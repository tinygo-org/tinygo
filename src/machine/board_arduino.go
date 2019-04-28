// +build avr,arduino

package machine

const CPU_FREQUENCY = 16000000

// LED on the Arduino
const LED Pin = 13

// ADC on the Arduino
const (
	ADC0 Pin = 0
	ADC1 Pin = 1
	ADC2 Pin = 2
	ADC3 Pin = 3
	ADC4 Pin = 4 // Used by TWI for SDA
	ADC5 Pin = 5 // Used by TWI for SCL
)

// UART pins
const (
	UART_TX_PIN Pin = 1
	UART_RX_PIN Pin = 0
)
