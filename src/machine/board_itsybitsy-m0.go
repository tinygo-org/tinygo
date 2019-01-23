// +build sam,atsamd21g18a,itsybitsy_m0

package machine

// GPIO Pins
const (
	D0  = 11 // UART0 RX: SERCOM0/PAD[3]
	D1  = 10 // UART0 TX: SERCOM0/PAD[2]
	D2  = 14
	D3  = 9
	D4  = 8
	D5  = 15
	D6  = 20
	D7  = 21
	D8  = 6
	D9  = 7
	D10 = 18 // UART1 TX(1): SERCOM1/PAD[2] can be used for UART1 TX
	D11 = 16 // UART1 TX(2): SERCOM1/PAD[0] can be used for UART1 TX
	D12 = 19
	D13 = 17
)

const (
	LED = D13
)

// UART0 pins
const (
	UART_TX_PIN = D1
	UART_RX_PIN = D0
)

// I2C pins
const (
	SDA_PIN = 22 // // SDA: SERCOM3/PAD[0]
	SCL_PIN = 23 // // SCL: SERCOM3/PAD[1]
)
