// +build lgt92

package machine

import (
	"device/stm32"
	"runtime/interrupt"
)

const (
	LED1 = PA12
	LED2 = PA8
	LED3 = PA11

	LED_RED   = LED1
	LED_BLUE  = LED2
	LED_GREEN = LED3

	BUTTON = PB14

	// LG GPS module
	GPS_STANDBY_PIN = PB3
	GPS_RESET_PIN   = PB4
	GPS_POWER_PIN   = PB5

	// LORA
	RFM95_DIO0_PIN = PC13

	//TinyGo UART is MCU LPUSART1
	UART_RX_PIN = PA13
	UART_TX_PIN = PA14

	//TinyGo UART1 is MCU USART1
	UART1_RX_PIN = PB6
	UART1_TX_PIN = PB7
)

var (

	// Console UART (LPUSART1)
	UART0 = UART{
		Buffer:          NewRingBuffer(),
		Bus:             stm32.LPUSART1,
		AltFuncSelector: 6,
	}

	// Gps UART
	UART1 = UART{
		Buffer:          NewRingBuffer(),
		Bus:             stm32.USART1,
		AltFuncSelector: 0,
	}
)

// set up RX IRQ handler. Follow similar pattern for other UARTx instances
func init() {

	// Enable GPIOs
	stm32.RCC.IOPENR.SetBits(stm32.RCC_IOPENR_IOPAEN)
	stm32.RCC.IOPENR.SetBits(stm32.RCC_IOPENR_IOPBEN)
	stm32.RCC.IOPENR.SetBits(stm32.RCC_IOPENR_IOPCEN)

	// Enable NVM (eeprom)
	stm32.RCC.AHBENR.SetBits(stm32.RCC_AHBENR_MIFEN)

	// Enable Leds GPIO output
	LED_RED.Configure(PinConfig{Mode: PinOutput})
	LED_GREEN.Configure(PinConfig{Mode: PinOutput})
	LED_BLUE.Configure(PinConfig{Mode: PinOutput})

	LED_RED.Low()
	LED_GREEN.Low()
	LED_BLUE.Low()

	// Enable UART0 (MCU LPUART1) Interrupt
	// Enable UART1 (MCU UART1) Interrupt
	UART0.Interrupt = interrupt.New(stm32.IRQ_AES_RNG_LPUART1, UART0.handleInterrupt)
	UART1.Interrupt = interrupt.New(stm32.IRQ_USART1, UART1.handleInterrupt)

}

// SPI pins
const (
	SPI1_SCK_PIN = PA5
	SPI1_SDI_PIN = PA6
	SPI1_SDO_PIN = PA7
	SPI0_SCK_PIN = SPI1_SCK_PIN
	SPI0_SDI_PIN = SPI1_SDI_PIN
	SPI0_SDO_PIN = SPI1_SDO_PIN
)

// MEMs accelerometer
const (
	MEMS_ACCEL_CS   = PE3
	MEMS_ACCEL_INT1 = PE0
	MEMS_ACCEL_INT2 = PE1
)

// Since the first interface is named SPI1, both SPI0 and SPI1 refer to SPI1.
// TODO: implement SPI2 and SPI3.
var (
	SPI0 = SPI{
		Bus: stm32.SPI1,
		//		AltFuncSelector: stm32.AF5_SPI1_SPI2,
	}
	SPI1 = &SPI0
)
