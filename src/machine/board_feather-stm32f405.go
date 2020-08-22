// +build feather_stm32f405

package machine

import (
	"device/stm32"
	"runtime/interrupt"
)

<<<<<<< HEAD
=======
// MCU port pins
const (
	PA0  = portA + 0
	PA1  = portA + 1
	PA2  = portA + 2
	PA3  = portA + 3
	PA4  = portA + 4
	PA5  = portA + 5
	PA6  = portA + 6
	PA7  = portA + 7
	PA8  = portA + 8
	PA9  = portA + 9
	PA10 = portA + 10
	PA11 = portA + 11
	PA12 = portA + 12
	PA13 = portA + 13
	PA14 = portA + 14
	PA15 = portA + 15

	PB0  = portB + 0
	PB1  = portB + 1
	PB2  = portB + 2
	PB3  = portB + 3
	PB4  = portB + 4
	PB5  = portB + 5
	PB6  = portB + 6
	PB7  = portB + 7
	PB8  = portB + 8
	PB9  = portB + 9
	PB10 = portB + 10
	PB11 = portB + 11
	PB12 = portB + 12
	PB13 = portB + 13
	PB14 = portB + 14
	PB15 = portB + 15

	PC0  = portC + 0
	PC1  = portC + 1
	PC2  = portC + 2
	PC3  = portC + 3
	PC4  = portC + 4
	PC5  = portC + 5
	PC6  = portC + 6
	PC7  = portC + 7
	PC8  = portC + 8
	PC9  = portC + 9
	PC10 = portC + 10
	PC11 = portC + 11
	PC12 = portC + 12
	PC13 = portC + 13
	PC14 = portC + 14
	PC15 = portC + 15

	PD0  = portD + 0
	PD1  = portD + 1
	PD2  = portD + 2
	PD3  = portD + 3
	PD4  = portD + 4
	PD5  = portD + 5
	PD6  = portD + 6
	PD7  = portD + 7
	PD8  = portD + 8
	PD9  = portD + 9
	PD10 = portD + 10
	PD11 = portD + 11
	PD12 = portD + 12
	PD13 = portD + 13
	PD14 = portD + 14
	PD15 = portD + 15

	PE0  = portE + 0
	PE1  = portE + 1
	PE2  = portE + 2
	PE3  = portE + 3
	PE4  = portE + 4
	PE5  = portE + 5
	PE6  = portE + 6
	PE7  = portE + 7
	PE8  = portE + 8
	PE9  = portE + 9
	PE10 = portE + 10
	PE11 = portE + 11
	PE12 = portE + 12
	PE13 = portE + 13
	PE14 = portE + 14
	PE15 = portE + 15

	PH0 = portH + 0
	PH1 = portH + 1
)

>>>>>>> add basic support for feather-stm32f405
const (
	NUM_DIGITAL_IO_PINS = 39
	NUM_ANALOG_IO_PINS  = 7
)

// Pinout
const (
	// Arduino pin = MCU port pin // primary functions (alternate functions)
	D0  = PB11 // USART3 RX, PWM TIM2_CH4 (I2C2 SDA)
	D1  = PB10 // USART3 TX, PWM TIM2_CH3 (I2C2 SCL, I2S2 BCK)
	D2  = PB3  // GPIO, SPI3 FLASH SCK
	D3  = PB4  // GPIO, SPI3 FLASH MISO
	D4  = PB5  // GPIO, SPI3 FLASH MOSI
	D5  = PC7  // GPIO, PWM TIM3_CH2 (USART6 RX, I2S3 MCK)
	D6  = PC6  // GPIO, PWM TIM3_CH1 (USART6 TX, I2S2 MCK)
	D7  = PA15 // GPIO, SPI3 FLASH CS
	D8  = PC0  // GPIO, Neopixel
	D9  = PB8  // GPIO, PWM TIM4_CH3 (CAN1 RX, I2C1 SCL)
	D10 = PB9  // GPIO, PWM TIM4_CH4 (CAN1 TX, I2C1 SDA, I2S2 WSL)
	D11 = PC3  // GPIO (I2S2 SD, SPI2 MOSI)
	D12 = PC2  // GPIO (I2S2ext SD, SPI2 MISO)
	D13 = PC1  // GPIO, Builtin LED
	D14 = PB7  // I2C1 SDA, PWM TIM4_CH2 (USART1 RX)
	D15 = PB6  // I2C1 SCL, PWM TIM4_CH1 (USART1 TX, CAN2 TX)
	D16 = PA4  // A0 (DAC OUT1)
	D17 = PA5  // A1 (DAC OUT2, SPI1 SCK)
	D18 = PA6  // A2, PWM TIM3_CH1 (SPI1 MISO)
	D19 = PA7  // A3, PWM TIM3_CH2 (SPI1 MOSI)
	D20 = PC4  // A4
	D21 = PC5  // A5
	D22 = PA3  // A6
	D23 = PB13 // SPI2 SCK, PWM TIM1_CH1N (I2S2 BCK, CAN2 TX)
	D24 = PB14 // SPI2 MISO, PWM TIM1_CH2N (I2S2ext SD)
	D25 = PB15 // SPI2 MOSI, PWM TIM1_CH3N (I2S2 SD)
	D26 = PC8  // SDIO
	D27 = PC9  // SDIO
	D28 = PC10 // SDIO
	D29 = PC11 // SDIO
	D30 = PC12 // SDIO
	D31 = PD2  // SDIO
	D32 = PB12 // SD Detect
	D33 = PC14 // OSC32
	D34 = PC15 // OSC32
	D35 = PA11 // USB D+
	D36 = PA12 // USB D-
	D37 = PA13 // SWDIO
	D38 = PA14 // SWCLK
)

// Analog pins
const (
	A0 = D16 // ADC12 IN4
	A1 = D17 // ADC12 IN5
	A2 = D18 // ADC12 IN6
	A3 = D19 // ADC12 IN7
	A4 = D20 // ADC12 IN14
	A5 = D21 // ADC12 IN15
	A6 = D22 // VBAT
)

// Pretty lights
const (
	LED          = LED_BUILTIN
	LED_BUILTIN  = LED_RED
	LED_RED      = D13
	LED_NEOPIXEL = D8
)

// UART pins
const (
	NUM_UART_INTERFACES = 3

	UART_RX_PIN = UART1_RX_PIN
	UART_TX_PIN = UART1_TX_PIN

	UART0_RX_PIN = D0
	UART0_TX_PIN = D1
	UART1_RX_PIN = UART0_RX_PIN
	UART1_TX_PIN = UART0_TX_PIN

	UART2_RX_PIN = D5
	UART2_TX_PIN = D6

	UART3_RX_PIN = D14
	UART3_TX_PIN = D15
)

var (
	// TBD: why do UART0 and UART1 have different types (struct vs reference)?
	UART0 = UART{
		Buffer:          NewRingBuffer(),
		Bus:             stm32.USART3,
		AltFuncSelector: stm32.AF7_USART1_2_3,
	}
	UART1 = &UART0

//	UART2 = &UART{
//		Buffer:          NewRingBuffer(),
//		Bus:             stm32.USART6,
//		AltFuncSelector: stm32.AF8_USART4_5_6,
//	}
//	UART3 = &UART{
//		Buffer:          NewRingBuffer(),
//		Bus:             stm32.USART1,
//		AltFuncSelector: stm32.AF7_USART1_2_3,
//	}
)

// set up RX IRQ handler. Follow similar pattern for other UARTx instances
func init() {
	UART0.Interrupt = interrupt.New(stm32.IRQ_USART3, UART0.handleInterrupt)
	//UART2.Interrupt = interrupt.New(stm32.IRQ_USART6, UART2.handleInterrupt)
	//UART3.Interrupt = interrupt.New(stm32.IRQ_USART1, UART3.handleInterrupt)
}

// SPI pins
const (
	NUM_SPI_INTERFACES = 3

	SPI_SCK_PIN = SPI1_SCK_PIN
	SPI_SDI_PIN = SPI1_SDI_PIN
	SPI_SDO_PIN = SPI1_SDO_PIN

	SPI0_SCK_PIN = D23
	SPI0_SDI_PIN = D24
	SPI0_SDO_PIN = D25
	SPI1_SCK_PIN = SPI0_SCK_PIN
	SPI1_SDI_PIN = SPI0_SDI_PIN
	SPI1_SDO_PIN = SPI0_SDO_PIN

	SPI2_SCK_PIN = D2
	SPI2_SDI_PIN = D3
	SPI2_SDO_PIN = D4

	SPI3_SCK_PIN = D17
	SPI3_SDI_PIN = D18
	SPI3_SDO_PIN = D19
)

// Since the first interface is named SPI1, both SPI0 and SPI1 refer to SPI1.
// TODO: implement SPI2 and SPI3.
var (
	// TBD: why do SPI0 and SPI1 have different types (struct vs reference)?
	SPI0 = SPI{
		Bus:             stm32.SPI2,
		AltFuncSelector: stm32.AF5_SPI1_SPI2,
	}
	SPI1 = &SPI0

//	SPI2 = &SPI{
//		Bus:             stm32.SPI3,
//		AltFuncSelector: stm32.AF6_SPI3,
//	}
//	SPI3 = &SPI{
//		Bus:             stm32.SPI1,
//		AltFuncSelector: stm32.AF5_SPI1_SPI2,
//	}
)
