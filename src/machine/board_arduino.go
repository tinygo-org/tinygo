// +build avr,arduino

package machine

const CPU_FREQUENCY = 16000000

// LED on the Arduino
const LED = 13

// ADC on the Arduino
const (
	ADC0 = 0
	ADC1 = 1
	ADC2 = 2
	ADC3 = 3
	ADC4 = 4 // Used by TWI for SDA
	ADC5 = 5 // Used by TWI for SCL
)

// UART pins
const (
	UART_TX_PIN = 1
	UART_RX_PIN = 0
)
