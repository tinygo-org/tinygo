//go:build stm32l0x1 && !nucleol031k6

// This file is for the bare STM32L0x1 (not for any specific board that is based
// on the STM32L0x1).

package machine

const (
	I2C0_SCL_PIN = NoPin
	I2C0_SDA_PIN = NoPin

	UART_TX_PIN = NoPin
	UART_RX_PIN = NoPin

	SPI0_SDI_PIN = NoPin
	SPI0_SDO_PIN = NoPin
	SPI0_SCK_PIN = NoPin
)
