// +build sam,atsamd21,xiao

package machine

import (
	"device/sam"
	"runtime/interrupt"
)

// used to reset into bootloader
const RESET_MAGIC_VALUE = 0xf01669ef

// GPIO Pins
const (
	D0  = PA02 // can be used for PWM or DAC
	D1  = PA04 // PWM available
	D2  = PA10 // PWM available
	D3  = PA11 // PWM available
	D4  = PA08 // can be used for PWM or I2C SDA
	D5  = PA09 // can be used for PWM or I2C SCL
	D6  = PB08 // can be used for PWM or UART1 TX
	D7  = PB09 // can be used for PWM or UART1 RX
	D8  = PA07 // can be used for PWM or SPI SCK
	D9  = PA05 // can be used for PWM or SPI MISO
	D10 = PA06 // can be used for PWM or SPI MOSI
)

// Analog pins
const (
	A0  = PA02 // ADC/AIN[0]
	A1  = PA04 // ADC/AIN[4]
	A2  = PA10 // ADC/AIN[18]
	A3  = PA11 // ADC/AIN[19]
	A4  = PA08 // ADC/AIN[16]
	A5  = PA09 // ADC/AIN[17]
	A6  = PB08 // ADC/AIN[2]
	A7  = PB09 // ADC/AIN[3]
	A8  = PA07 // ADC/AIN[7]
	A9  = PA05 // ADC/AIN[6]
	A10 = PA06 // ADC/AIN[5]
)

const (
	LED         = PA17
	PIN_LED_13  = PA17
	PIN_LED     = PA17
	LED_BUILTIN = PA17

	PIN_LED_RXL = PA18
	PIN_LED_TXL = PA19
	PIN_LED2    = PIN_LED_RXL
	PIN_LED3    = PIN_LED_TXL
)

const (
	SWDIO = PA31
	SWCLK = PA30
)

// UART0 aka USBCDC pins
const (
	USBCDC_DM_PIN = PA24
	USBCDC_DP_PIN = PA25
)

// UART1 pins
const (
	UART_TX_PIN = D6
	UART_RX_PIN = D7
)

// UART1 on the Xiao
var (
	UART1 = UART{
		Buffer: NewRingBuffer(),
		Bus:    sam.SERCOM4_USART,
		SERCOM: 4,
	}
)

func init() {
	UART1.Interrupt = interrupt.New(sam.IRQ_SERCOM4, UART1.handleInterrupt)
}

// I2C pins
const (
	SDA_PIN = PA08 // SDA: SERCOM2/PAD[0]
	SCL_PIN = PA09 // SCL: SERCOM2/PAD[1]
)

// I2C on the Xiao
var (
	I2C0 = I2C{
		Bus:    sam.SERCOM2_I2CM,
		SERCOM: 2,
	}
)

// SPI pins
const (
	SPI0_SCK_PIN  = PA07 // SCK: SERCOM0/PAD[3]
	SPI0_MOSI_PIN = PA06 // MOSI: SERCOM0/PAD[2]
	SPI0_MISO_PIN = PA05 // MISO: SERCOM0/PAD[1]
)

// SPI on the Xiao
var (
	SPI0 = SPI{
		Bus:    sam.SERCOM0_SPI,
		SERCOM: 0,
	}
)

// I2S pins
const (
	I2S_SCK_PIN = PA10
	I2S_SD_PIN  = PA08
	I2S_WS_PIN  = NoPin // TODO: figure out what this is on Xiao
)

// I2S on the Xiao
var (
	I2S0 = I2S{Bus: sam.I2S}
)

// USB CDC identifiers
const (
	usb_STRING_PRODUCT      = "Seeed XIAO M0"
	usb_STRING_MANUFACTURER = "Seeed"
)

var (
	usb_VID uint16 = 0x2886
	usb_PID uint16 = 0x802F
)
