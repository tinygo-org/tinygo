//go:build feather_stm32f405

package machine

import (
	"device/stm32"
	"runtime/interrupt"
)

const (
	NUM_DIGITAL_IO_PINS = 39
	NUM_ANALOG_IO_PINS  = 7
)

// Digital pins
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

func init() {
	initLED()
	initUART()
	initSPI()
	initI2C()
}

// -- LEDs ---------------------------------------------------------------------

const (
	NUM_BOARD_LED      = 1
	NUM_BOARD_NEOPIXEL = 1

	LED_RED      = D13
	LED_NEOPIXEL = D8
	LED_BUILTIN  = LED_RED
	LED          = LED_BUILTIN
	WS2812       = D8
)

func initLED() {}

// -- UART ---------------------------------------------------------------------

const (
	// #===========#==========#==============#============#=======#=======#
	// | Interface | Hardware |  Bus(Freq)   | RX/TX Pins | AltFn | Alias |
	// #===========#==========#==============#============#=======#=======#
	// |   UART1   |  USART3  | APB1(42 MHz) |   D0/D1    |   7   |   ~   |
	// |   UART2   |  USART6  | APB2(84 MHz) |   D5/D6    |   8   |   ~   |
	// |   UART3   |  USART1  | APB2(84 MHz) |  D14/D15   |   7   |   ~   |
	// | --------- | -------- | ------------ | ---------- | ----- | ----- |
	// |   UART0   |  USART3  | APB1(42 MHz) |   D0/D1    |   7   | UART1 |
	// #===========#==========#==============#============#=======#=======#
	NUM_UART_INTERFACES = 3

	UART1_RX_PIN = D0 // UART1 = hardware: USART3
	UART1_TX_PIN = D1 //

	UART2_RX_PIN = D5 // UART2 = hardware: USART6
	UART2_TX_PIN = D6 //

	UART3_RX_PIN = D14 // UART3 = hardware: USART1
	UART3_TX_PIN = D15 //

	UART0_RX_PIN = UART1_RX_PIN // UART0 = alias: UART1
	UART0_TX_PIN = UART1_TX_PIN //

	UART_RX_PIN = UART0_RX_PIN // default/primary UART pins
	UART_TX_PIN = UART0_TX_PIN //
)

var (
	UART1  = &_UART1
	_UART1 = UART{
		Buffer:            NewRingBuffer(),
		Bus:               stm32.USART3,
		TxAltFuncSelector: AF7_USART1_2_3,
		RxAltFuncSelector: AF7_USART1_2_3,
	}
	UART2  = &_UART2
	_UART2 = UART{
		Buffer:            NewRingBuffer(),
		Bus:               stm32.USART6,
		TxAltFuncSelector: AF8_USART4_5_6,
		RxAltFuncSelector: AF8_USART4_5_6,
	}
	UART3  = &_UART3
	_UART3 = UART{
		Buffer:            NewRingBuffer(),
		Bus:               stm32.USART1,
		TxAltFuncSelector: AF7_USART1_2_3,
		RxAltFuncSelector: AF7_USART1_2_3,
	}
	DefaultUART = UART1
)

func initUART() {
	UART1.Interrupt = interrupt.New(stm32.IRQ_USART3, _UART1.handleInterrupt)
	UART2.Interrupt = interrupt.New(stm32.IRQ_USART6, _UART2.handleInterrupt)
	UART3.Interrupt = interrupt.New(stm32.IRQ_USART1, _UART3.handleInterrupt)
}

// -- SPI ----------------------------------------------------------------------

const (
	// #===========#==========#==============#==================#=======#=======#
	// | Interface | Hardware |  Bus(Freq)   | SCK/SDI/SDO Pins | AltFn | Alias |
	// #===========#==========#==============#==================#=======#=======#
	// |   SPI1    |   SPI2   | APB1(42 MHz) |    D23/D24/D25   |   5   |   ~   |
	// |   SPI2    |   SPI3   | APB1(42 MHz) |     D2/D3/D4     |   6   |   ~   |
	// |   SPI3    |   SPI1   | APB2(84 MHz) |    D17/D18/D19   |   5   |   ~   |
	// | --------- | -------- | ------------ | ---------------- | ----- | ----- |
	// |   SPI0    |   SPI2   | APB1(42 MHz) |    D23/D24/D25   |   5   | SPI1  |
	// #===========#==========#==============#==================#=======#=======#
	NUM_SPI_INTERFACES = 3

	SPI1_SCK_PIN = D23 //
	SPI1_SDI_PIN = D24 // SPI1 = hardware: SPI2
	SPI1_SDO_PIN = D25 //

	SPI2_SCK_PIN = D2 //
	SPI2_SDI_PIN = D3 // SPI2 = hardware: SPI3
	SPI2_SDO_PIN = D4 //

	SPI3_SCK_PIN = D17 //
	SPI3_SDI_PIN = D18 // SPI3 = hardware: SPI1
	SPI3_SDO_PIN = D19 //

	SPI0_SCK_PIN = SPI1_SCK_PIN //
	SPI0_SDI_PIN = SPI1_SDI_PIN // SPI0 = alias: SPI1
	SPI0_SDO_PIN = SPI1_SDO_PIN //

	SPI_SCK_PIN = SPI0_SCK_PIN //
	SPI_SDI_PIN = SPI0_SDI_PIN // default/primary SPI pins
	SPI_SDO_PIN = SPI0_SDO_PIN //
)

var (
	SPI1 = SPI{
		Bus:             stm32.SPI2,
		AltFuncSelector: AF5_SPI1_SPI2,
	}
	SPI2 = SPI{
		Bus:             stm32.SPI3,
		AltFuncSelector: AF6_SPI3,
	}
	SPI3 = SPI{
		Bus:             stm32.SPI1,
		AltFuncSelector: AF5_SPI1_SPI2,
	}
	SPI0 = SPI1
)

func initSPI() {}

// -- I2C ----------------------------------------------------------------------

const (
	// #===========#==========#==============#==============#=======#=======#
	// | Interface | Hardware |  Bus(Freq)   | SDA/SCL Pins | AltFn | Alias |
	// #===========#==========#==============#==============#=======#=======#
	// |   I2C1    |   I2C1   | APB1(42 MHz) |   D14/D15    |   4   |   ~   |
	// |   I2C2    |   I2C2   | APB1(42 MHz) |    D0/D1     |   4   |   ~   |
	// |   I2C3    |   I2C1   | APB1(42 MHz) |    D9/D10    |   4   |   ~   |
	// | --------- | -------- | ------------ | ------------ | ----- | ----- |
	// |   I2C0    |   I2C1   | APB1(42 MHz) |   D14/D15    |   4   | I2C1  |
	// #===========#==========#==============#==============#=======#=======#
	NUM_I2C_INTERFACES = 3

	I2C1_SDA_PIN = D14 // I2C1 = hardware: I2C1
	I2C1_SCL_PIN = D15 //

	I2C2_SDA_PIN = D0 // I2C2 = hardware: I2C2
	I2C2_SCL_PIN = D1 //

	I2C3_SDA_PIN = D9  // I2C3 = hardware: I2C1
	I2C3_SCL_PIN = D10 //   (interface duplicated on second pair of pins)

	I2C0_SDA_PIN = I2C1_SDA_PIN // I2C0 = alias: I2C1
	I2C0_SCL_PIN = I2C1_SCL_PIN //

	I2C_SDA_PIN = I2C0_SDA_PIN // default/primary I2C pins
	I2C_SCL_PIN = I2C0_SCL_PIN //

	SDA_PIN = I2C0_SDA_PIN
	SCL_PIN = I2C0_SCL_PIN
)

var (
	I2C1 = &I2C{
		Bus:             stm32.I2C1,
		AltFuncSelector: AF4_I2C1_2_3,
	}
	I2C2 = &I2C{
		Bus:             stm32.I2C2,
		AltFuncSelector: AF4_I2C1_2_3,
	}
	I2C3 = &I2C{
		Bus:             stm32.I2C1,
		AltFuncSelector: AF4_I2C1_2_3,
	}
	I2C0 = I2C1
)

func initI2C() {}
