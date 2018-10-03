// +build nrf,pca10040

package machine

// LEDs on the PCA10040 (nRF52832 dev board)
const (
	LED  = LED1
	LED1 = 17
	LED2 = 18
	LED3 = 19
	LED4 = 20
)

// Buttons on the PCA10040 (nRF52832 dev board)
const (
	BUTTON  = BUTTON1
	BUTTON1 = 13
	BUTTON2 = 14
	BUTTON3 = 15
	BUTTON4 = 16
)

// UART pins for NRF52840-DK
const (
	UART_TX_PIN = 6
	UART_RX_PIN = 8
)

// ADC pins
const (
	ADC0 = 3
	ADC1 = 4
	ADC2 = 28
	ADC3 = 29
	ADC4 = 30
	ADC5 = 31
)

// I2C pins
const (
	SDA_PIN = 26
	SCL_PIN = 27
)
