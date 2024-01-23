//go:build mksnanov3

// The MKS Robin Nano V3.X board.
// Documented at https://github.com/makerbase-mks/MKS-Robin-Nano-V3.X.

package machine

import (
	"device/stm32"
	"runtime/interrupt"
)

// LED is also wired to the SD card card detect (CD) pin.
const LED = PD12

// UART pins
const (
	UART_TX_PIN = PB10
	UART_RX_PIN = PB11
)

// EXP1 and EXP2 expansion ports for connecting
// the MKS TS35 V2.0 expansion board.
const (
	BEEPER = EXP1_1

	// LCD pins.
	LCD_DC        = EXP1_8
	LCD_CS        = EXP1_7
	LCD_RS        = EXP1_4
	LCD_BACKLIGHT = EXP1_3

	// Touch pins. Note that some pins are shared with the
	// LCD SPI1 interface.
	TOUCH_CLK  = EXP2_2
	TOUCH_CS   = EXP1_5
	TOUCH_DIN  = EXP2_6
	TOUCH_DOUT = EXP2_1
	TOUCH_IRQ  = EXP1_6

	BUTTON         = BUTTON_JOG
	BUTTON_JOG     = EXP1_2
	BUTTON_JOG_CCW = EXP2_3
	BUTTON_JOG_CW  = EXP2_5

	EXP1_1 = PC5
	EXP1_2 = PE13
	EXP1_3 = PD13
	EXP1_4 = PC6
	EXP1_5 = PE14
	EXP1_6 = PE15
	EXP1_7 = PD11
	EXP1_8 = PD10

	EXP2_1 = PA6
	EXP2_2 = PA5
	EXP2_3 = PE8
	EXP2_4 = PE10
	EXP2_5 = PE11
	EXP2_6 = PA7
	EXP2_7 = PE12
)

var (
	UART3  = &_UART3
	_UART3 = UART{
		Buffer:            NewRingBuffer(),
		Bus:               stm32.USART3,
		TxAltFuncSelector: AF7_USART1_2_3,
		RxAltFuncSelector: AF7_USART1_2_3,
	}
	DefaultUART = UART3
)

// set up RX IRQ handler. Follow similar pattern for other UARTx instances
func init() {
	UART3.Interrupt = interrupt.New(stm32.IRQ_USART3, _UART3.handleInterrupt)
}

// SPI pins
const (
	SPI1_SCK_PIN = EXP2_2
	SPI1_SDI_PIN = EXP2_1
	SPI1_SDO_PIN = EXP2_6
	SPI0_SCK_PIN = SPI1_SCK_PIN
	SPI0_SDI_PIN = SPI1_SDI_PIN
	SPI0_SDO_PIN = SPI1_SDO_PIN
)

// Since the first interface is named SPI1, both SPI0 and SPI1 refer to SPI1.
var (
	SPI0 = SPI{
		Bus:             stm32.SPI1,
		AltFuncSelector: AF5_SPI1_SPI2,
	}
	SPI1 = &SPI0
)

const (
	I2C0_SCL_PIN = PB6
	I2C0_SDA_PIN = PB7
)

var (
	I2C0 = &I2C{
		Bus:             stm32.I2C1,
		AltFuncSelector: AF4_I2C1_2_3,
	}
)

// Motor control pins.
const (
	X_ENABLE = PE4
	X_STEP   = PE3
	X_DIR    = PE2
	X_DIAG   = PA15
	X_UART   = PD5

	Y_ENABLE = PE1
	Y_STEP   = PE0
	Y_DIR    = PB9
	Y_DIAG   = PD2
	Y_UART   = PD7

	Z_ENABLE = PB8
	Z_STEP   = PB5
	Z_DIR    = PB4
	Z_DIAG   = PC8
	Z_UART   = PD4

	E0_ENABLE = PB3
	E0_STEP   = PD6
	E0_DIR    = PD3
	E0_DIAG   = PC4
	E0_UART   = PD9

	E1_ENABLE = PA3
	E1_STEP   = PD15
	E1_DIR    = PA1
	E1_DIAG   = PE7
	E1_UART   = PD8
)
