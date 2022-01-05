//go:build m5stamp_c3
// +build m5stamp_c3

package machine

const (
	IO0  Pin = 0
	IO1  Pin = 1
	IO2  Pin = 2
	IO3  Pin = 3
	IO4  Pin = 4
	IO5  Pin = 5
	IO6  Pin = 6
	IO7  Pin = 7
	IO8  Pin = 8
	IO9  Pin = 9
	IO10 Pin = 10
	IO11 Pin = 11
	IO12 Pin = 12
	IO13 Pin = 13
	IO14 Pin = 14
	IO15 Pin = 15
	IO16 Pin = 16
	IO17 Pin = 17
	IO18 Pin = 18
	IO19 Pin = 19
	IO20 Pin = 20
	IO21 Pin = 21

	XTAL_32K_P = IO0
	XTAL_32K_N = IO1
	MTMS       = IO4
	MTDI       = IO5
	MTCK       = IO6
	MTDO       = IO7
	VDD_SPI    = IO11
	SPIHD      = IO12
	SPISP      = IO13
	SPICS0     = IO14
	SPICLK     = IO15
	SPID       = IO16
	SPIQ       = IO17
	U0RXD      = IO20
	U0TXD      = IO21
)

const (
	WS2812 = IO2
)
