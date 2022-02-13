//go:build lgt92
// +build lgt92

package machine

import (
	"runtime/interrupt"
	"tinygo.org/x/device/stm32"
)

const (
	LED1 = PA12
	LED2 = PA8
	LED3 = PA11

	LED_RED   = LED1
	LED_BLUE  = LED2
	LED_GREEN = LED3

	// Default led
	LED = LED1

	BUTTON = PB14

	// LG GPS module
	GPS_STANDBY_PIN = PB3
	GPS_RESET_PIN   = PB4
	GPS_POWER_PIN   = PB5

	MEMS_ACCEL_CS   = PE3
	MEMS_ACCEL_INT1 = PE0
	MEMS_ACCEL_INT2 = PE1

	// SPI
	SPI1_SCK_PIN = PA5
	SPI1_SDI_PIN = PA6
	SPI1_SDO_PIN = PA7
	SPI0_SCK_PIN = SPI1_SCK_PIN
	SPI0_SDI_PIN = SPI1_SDI_PIN
	SPI0_SDO_PIN = SPI1_SDO_PIN

	// LORA RFM95 Radio
	RFM95_DIO0_PIN = PC13

	// TinyGo UART is MCU LPUSART1
	UART_RX_PIN = PA13
	UART_TX_PIN = PA14

	// TinyGo UART1 is MCU USART1
	UART1_RX_PIN = PB6
	UART1_TX_PIN = PB7

	// MPU9250 Nine-Axis (Gyro + Accelerometer + Compass)
	I2C0_SCL_PIN = PA9
	I2C0_SDA_PIN = PA10
)

var DefaultUART = UART0

var (

	// Console UART (LPUSART1)
	UART0  = &_UART0
	_UART0 = UART{
		Buffer:            NewRingBuffer(),
		Bus:               stm32.LPUART1,
		TxAltFuncSelector: 6,
		RxAltFuncSelector: 6,
	}

	// Gps UART
	UART1  = &_UART1
	_UART1 = UART{
		Buffer:            NewRingBuffer(),
		Bus:               stm32.USART1,
		TxAltFuncSelector: 0,
		RxAltFuncSelector: 0,
	}

	// MPU9250 Nine-Axis (Gyro + Accelerometer + Compass)
	I2C1 = &I2C{
		Bus:             stm32.I2C1,
		AltFuncSelector: 6,
	}
	I2C0 = I2C1

	// SPI
	SPI0 = SPI{
		Bus: stm32.SPI1,
	}
	SPI1 = &SPI0
)

func init() {
	// Enable UARTs Interrupts
	UART0.Interrupt = interrupt.New(stm32.IRQ_AES_RNG_LPUART1, _UART0.handleInterrupt)
	UART1.Interrupt = interrupt.New(stm32.IRQ_USART1, _UART1.handleInterrupt)
}
