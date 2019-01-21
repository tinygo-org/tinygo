// +build sam,atsamd21g18a,itsybitsy_m0

package machine

// GPIO Pins
const (
	D0  = 11 // RX: SERCOM0/PAD[3]
	D1  = 10 // TX: SERCOM0/PAD[2]
	D2  = 14
	D3  = 9
	D4  = 8
	D5  = 15
	D6  = 20
	D7  = 21
	D8  = 6
	D9  = 7
	D10 = 18
	D11 = 16
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
