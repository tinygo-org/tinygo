//go:build gnse

package machine

import (
	"device/stm32"
	"runtime/interrupt"
)

const (
	LED_RED   = PB5
	LED_GREEN = PB6
	LED_BLUE  = PB7
	LED1      = LED_RED   // Red
	LED2      = LED_GREEN // Green
	LED3      = LED_BLUE  // Blue
	LED       = LED_GREEN // Default

	BUTTON    = PB3
	BUZZER    = PA15
	VBATT_ADC = PB2
	SENSOR_EN = PB12
	FLASH_EN  = PC13

	// SPI0
	SPI0_NSS_PIN = PA4
	SPI0_SCK_PIN = PA5
	SPI0_SDO_PIN = PA6
	SPI0_SDI_PIN = PA7

	//MCU USART2
	UART2_RX_PIN = PA3
	UART2_TX_PIN = PA2

	// DEFAULT USART
	UART_RX_PIN = UART2_RX_PIN
	UART_TX_PIN = UART2_TX_PIN

	// I2C1 pins
	// I2C1 is connected to Flash, Accelerometer, Env. Sensor, Crypto Element)
	I2C1_SCL_PIN  = PA9
	I2C1_SDA_PIN  = PA10
	I2C1_ALT_FUNC = 4

	// I2C2 pins
	// I2C2 is expansion J10 QWIIC Connector
	I2C2_SCL_PIN  = PA12
	I2C2_SDA_PIN  = PA11
	I2C2_ALT_FUNC = 4

	// I2C0 alias for I2C1
	I2C0_SDA_PIN = I2C1_SDA_PIN
	I2C0_SCL_PIN = I2C1_SCL_PIN
)

var (
	// STM32 UART2 is connected to the embedded STLINKV3 Virtual Com Port
	UART0  = &_UART0
	_UART0 = UART{
		UARTCommon:        NewUARTCommon(),
		Bus:               stm32.USART2,
		TxAltFuncSelector: 7,
		RxAltFuncSelector: 7,
	}

	DefaultUART = UART0

	// I2C Busses
	I2C1 = &I2C{
		Bus:             stm32.I2C1,
		AltFuncSelector: I2C1_ALT_FUNC,
	}
	I2C2 = &I2C{
		Bus:             stm32.I2C2,
		AltFuncSelector: I2C2_ALT_FUNC,
	}
	I2C0 = I2C1

	// SPI
	SPI3 = SPI{
		Bus: stm32.SPI3,
	}
)

func init() {
	// Enable UARTs Interrupts
	UART0.Interrupt = interrupt.New(stm32.IRQ_USART2, _UART0.handleInterrupt)
}
