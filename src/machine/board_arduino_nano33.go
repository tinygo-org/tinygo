// +build sam,atsamd21,arduino_nano33

// This contains the pin mappings for the Arduino Nano33 IoT board.
//
// For more information, see: https://store.arduino.cc/nano-33-iot
//
package machine

import "device/sam"

// used to reset into bootloader
const RESET_MAGIC_VALUE = 0x07738135

// GPIO Pins
const (
	RX0 Pin = PB23 // UART2 RX
	TX1 Pin = PB22 // UART2 TX

	D2 Pin = PB10 // PWM available
	D3 Pin = PB11 // PWM available
	D4 Pin = PA07
	D5 Pin = PA05 // PWM available
	D6 Pin = PA04 // PWM available
	D7 Pin = PA06

	D8  Pin = PA18
	D9  Pin = PA20 // PWM available
	D10 Pin = PA21 // PWM available
	D11 Pin = PA16 // PWM available
	D12 Pin = PA19 // PWM available

	D13 Pin = PA17
)

// Analog pins
const (
	A0 Pin = PA02 // ADC/AIN[0]
	A1 Pin = PB02 // ADC/AIN[10]
	A2 Pin = PA11 // ADC/AIN[19]
	A3 Pin = PA10 // ADC/AIN[18],
	A4 Pin = PB08 // ADC/AIN[2], SCL:  SERCOM2/PAD[1]
	A5 Pin = PB09 // ADC/AIN[3], SDA:  SERCOM2/PAD[1]
	A6 Pin = PA09 // ADC/AIN[17]
	A7 Pin = PB03 // ADC/AIN[11]
)

const (
	LED = D13
)

// NINA-W102 Pins

const (
	NINA_MOSI   Pin = PA12
	NINA_MISO   Pin = PA13
	NINA_CS     Pin = PA14
	NINA_SCK    Pin = PA15
	NINA_GPIO0  Pin = PA27
	NINA_RESETN Pin = PA08
	NINA_ACK    Pin = PA28
)

// UART0 aka USBCDC pins
const (
	USBCDC_DM_PIN Pin = PA24
	USBCDC_DP_PIN Pin = PA25
)

// UART1 on the Arduino Nano 33 connects to the onboard NINA-W102 WiFi chip.
var (
	UART1 = UART{Bus: sam.SERCOM5_USART,
		Buffer: NewRingBuffer(),
		Mode:   PinSERCOMAlt,
		IRQVal: sam.IRQ_SERCOM5,
	}
)

// UART1 pins
const (
	UART_TX_PIN Pin = PA22
	UART_RX_PIN Pin = PA23
)

//go:export SERCOM5_IRQHandler
func handleUART1() {
	defaultUART1Handler()
}

// UART2 on the Arduino Nano 33 connects to the normal TX/RX pins.
var (
	UART2 = UART{Bus: sam.SERCOM3_USART,
		Buffer: NewRingBuffer(),
		Mode:   PinSERCOMAlt,
		IRQVal: sam.IRQ_SERCOM3,
	}
)

//go:export SERCOM3_IRQHandler
func handleUART2() {
	// should reset IRQ
	UART2.Receive(byte((UART2.Bus.DATA.Get() & 0xFF)))
	UART2.Bus.INTFLAG.SetBits(sam.SERCOM_USART_INTFLAG_RXC)
}

// I2C pins
const (
	SDA_PIN Pin = A4 // SDA: SERCOM4/PAD[1]
	SCL_PIN Pin = A5 // SCL: SERCOM4/PAD[1]
)

// I2C on the Arduino Nano 33.
var (
	I2C0 = I2C{
		Bus:    sam.SERCOM4_I2CM,
		SERCOM: 4,
	}
)

// SPI pins
const (
	SPI0_SCK_PIN  Pin = A2 // SCK: SERCOM0/PAD[3]
	SPI0_MOSI_PIN Pin = A3 // MOSI: SERCOM0/PAD[2]
	SPI0_MISO_PIN Pin = A6 // MISO: SERCOM0/PAD[1]
)

// SPI on the Arduino Nano 33.
var (
	SPI0 = SPI{
		Bus:    sam.SERCOM0_SPI,
		SERCOM: 0,
	}
)

// I2S pins
const (
	I2S_SCK_PIN Pin = PA10
	I2S_SD_PIN  Pin = PA08
	I2S_WS_PIN      = NoPin // TODO: figure out what this is on Arduino Nano 33.
)

// I2S on the Arduino Nano 33.
var (
	I2S0 = I2S{Bus: sam.I2S}
)
