// +build hifive1b

package machine

const (
	D0  = P16
	D1  = P17
	D2  = P18
	D3  = P19 // Green LED/PWM
	D4  = P20 // PWM
	D5  = P21 // Blue LED/PWM
	D6  = P22 // Red LED/PWM
	D7  = P16
	D8  = NoPin // PWM?
	D9  = P01
	D10 = P02
	D11 = P03
	D12 = P04
	D13 = P05
	D14 = NoPin // not connected
	D15 = P09   // does not seem to work?
	D16 = P10   // PWM
	D17 = P11   // PWM
	D18 = P12   // SDA/PWM
	D19 = P13   // SDL/PWM
)

const (
	LED       = LED1
	LED1      = LED_RED
	LED2      = LED_GREEN
	LED3      = LED_BLUE
	LED_RED   = P22
	LED_GREEN = P19
	LED_BLUE  = P21
)

const (
	// TODO: figure out the pin numbers for these.
	UART_TX_PIN = NoPin
	UART_RX_PIN = NoPin
)
