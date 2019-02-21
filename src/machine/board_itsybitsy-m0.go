// +build sam,atsamd21,itsybitsy_m0

package machine

// GPIO Pins
const (
	D0  = 11 // UART0 RX
	D1  = 10 // UART0 TX
	D2  = 14
	D3  = 9  // PWM available
	D4  = 8  // PWM available
	D5  = 15 // PWM available
	D6  = 20 // PWM available
	D7  = 21 // PWM available
	D8  = 6  // PWM available
	D9  = 7  // PWM available
	D10 = 18 // can be used for PWM or UART1 TX
	D11 = 16 // can be used for PWM or UART1 RX
	D12 = 19 // PWM available
	D13 = 17 // PWM available
)

// Analog pins
const (
	A0 = 2 // ADC/AIN[0]
	// A1 = 8 // ADC/AIN[2] TODO: requires PORTB
	// A2 = 9 // ADC/AIN[3] TODO: requires PORTB
	A3 = 4 // ADC/AIN[4]
	A4 = 5 // ADC/AIN[5]
	//A5 = 2 // ADC/AIN[10] TODO: requires PORTB
)

const (
	LED = D13
)

// UART0 aka USBCDC pins
const (
	USBCDC_DM_PIN = 24
	USBCDC_DP_PIN = 25
)

// UART1 pins
const (
	UART_TX_PIN = D1
	UART_RX_PIN = D0
)

// I2C pins
const (
	SDA_PIN = 22 // SDA: SERCOM3/PAD[0]
	SCL_PIN = 23 // SCL: SERCOM3/PAD[1]
)
