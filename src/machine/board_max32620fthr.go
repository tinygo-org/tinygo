// +build max32620fthr

package machine

// https://github.com/MaximIntegratedMicros/arduino-max326xx/blob/master/arm/variants/max32620_fthr/variant.h
// https://www.digikey.com.au/htmldatasheets/production/2911699/0/0/1/MAX32620FTHR.pdf

const (
	A0 = AIN_0
	A1 = AIN_1
	A2 = AIN_2
	A3 = AIN_3
)

const (
	LED_RED   = P2_4
	LED_GREEN = P2_5
	LED_BLUE  = P2_6

	LED = LED_RED
)

// UART
const (
	UART0_TX_PIN = P0_1
	UART0_RX_PIN = P0_0

	UART1_TX_PIN = P2_1 // USBTX
	UART1_RX_PIN = P2_0 // USBRX

	UART2_TX_PIN = P3_1
	UART2_RX_PIN = P3_0
)
