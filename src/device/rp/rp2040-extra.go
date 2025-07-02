// Hand created file. DO NOT DELETE.
// Definitions that are missing in src/device/rp/rp2040.go generated from SVDs

//go:build rp && rp2040

package rp

// DMA: 2.5.3.1. System DREQ Table from RP2040 Datasheet
// https://datasheets.raspberrypi.com/rp2040/rp2040-datasheet.pdf
const (
	DREQ_PIO0_TX0   = 0
	DREQ_PIO0_TX1   = 1
	DREQ_PIO0_TX2   = 2
	DREQ_PIO0_TX3   = 3
	DREQ_PIO0_RX0   = 4
	DREQ_PIO0_RX1   = 5
	DREQ_PIO0_RX2   = 6
	DREQ_PIO0_RX3   = 7
	DREQ_PIO1_TX0   = 8
	DREQ_PIO1_TX1   = 9
	DREQ_PIO1_TX2   = 10
	DREQ_PIO1_TX3   = 11
	DREQ_PIO1_RX0   = 12
	DREQ_PIO1_RX1   = 13
	DREQ_PIO1_RX2   = 14
	DREQ_PIO1_RX3   = 15
	DREQ_SPI0_TX    = 16
	DREQ_SPI0_RX    = 17
	DREQ_SPI1_TX    = 18
	DREQ_SPI1_RX    = 19
	DREQ_UART0_TX   = 20
	DREQ_UART0_RX   = 21
	DREQ_UART1_TX   = 22
	DREQ_UART1_RX   = 23
	DREQ_PWM_WRAP0  = 24
	DREQ_PWM_WRAP1  = 25
	DREQ_PWM_WRAP2  = 26
	DREQ_PWM_WRAP3  = 27
	DREQ_PWM_WRAP4  = 28
	DREQ_PWM_WRAP5  = 29
	DREQ_PWM_WRAP6  = 30
	DREQ_PWM_WRAP7  = 31
	DREQ_I2C0_TX    = 32
	DREQ_I2C0_RX    = 33
	DREQ_I2C1_TX    = 34
	DREQ_I2C1_RX    = 35
	DREQ_ADC        = 36
	DREQ_XIP_STREAM = 37
	DREQ_XIP_SSITX  = 38
	DREQ_XIP_SSIRX  = 39
)
