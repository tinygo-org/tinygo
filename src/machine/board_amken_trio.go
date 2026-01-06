//go:build amken_trio

// RabbitPNP Toolhead Board
// MCU: STM32G0B1CBTx (LQFP48, 128KB Flash, 144KB RAM)

package machine

import (
	"device/stm32"
	"runtime/interrupt"
)

// Vacuum sensors (PWM input via TIM2)
const (
	VAC1 = PA0 // TIM2_CH1
	VAC2 = PA1 // TIM2_CH2
)

// Motor 1 pins (stepper driver)
const (
	M1_CS   = PC7
	M1_DIR  = PB13 // REFL
	M1_STEP = PB14 // REFR
	M1_ENN  = PA10 // Enable (active low)
)

// Motor 2 pins (stepper driver)
const (
	M2_CS   = PA9
	M2_DIR  = PB12 // REFL
	M2_STEP = PB11 // REFR
	M2_ENN  = PC6  // Enable (active low)
)

// Motor 3 pins (stepper driver)
const (
	M3_CS   = PB15
	M3_DIR  = PB2 // REFL
	M3_STEP = PB1 // REFR
	M3_ENN  = PA8 // Enable (active low)
)

// LED
const (
	LED         = LED1
	LED_BUILTIN = LED1
	LED1        = PB7
)

// Solenoid and Neopixel (PWM via TIM4)
const (
	SOLENOID = PB8 // TIM4_CH3
	NEOPIXEL = PB9 // TIM4_CH4
)

// Endstops
const (
	ENDSTOP_IN1 = PC13
	ENDSTOP_IN2 = PC14
)

// Magnetic sensor
const (
	MAG1 = PB3
)

// Accelerometer chip select (LIS2D on SPI2)
const (
	LIS2D_CS = PB5
)

// SPI1 pins (motor drivers)
const (
	SPI1_SCK_PIN = PA5
	SPI1_SDO_PIN = PA2 // MOSI
	SPI1_SDI_PIN = PA6 // MISO
	SPI0_SCK_PIN = SPI1_SCK_PIN
	SPI0_SDO_PIN = SPI1_SDO_PIN
	SPI0_SDI_PIN = SPI1_SDI_PIN
)

// SPI2 pins (accelerometer)
const (
	SPI2_SCK_PIN = PB10
	SPI2_SDO_PIN = PA4 // MOSI
	SPI2_SDI_PIN = PA3 // MISO
)

// I2C2 pins
const (
	I2C2_SCL_PIN = PA7
	I2C2_SDA_PIN = PB4
	I2C0_SCL_PIN = I2C2_SCL_PIN
	I2C0_SDA_PIN = I2C2_SDA_PIN
)

// FDCAN1 pins
const (
	CAN_RX = PD0
	CAN_TX = PD1
)

// USB pins
const (
	USB_DM = PA11
	USB_DP = PA12
)

// UART pins (not directly connected but required by machine package)
const (
	UART_TX_PIN = NoPin
	UART_RX_PIN = NoPin
)

var (
	// SPI1 for motor drivers
	SPI1 = &SPI{
		Bus:             stm32.SPI1,
		AltFuncSelector: AF0_SYSTEM,
	}
	SPI0 = SPI1

	// SPI2 for accelerometer
	SPI2 = &SPI{
		Bus:             stm32.SPI2,
		AltFuncSelector: AF1_TIM1_TIM2_TIM3_LPTIM1,
	}

	// I2C2
	I2C2 = &I2C{
		Bus:             stm32.I2C2,
		AltFuncSelector: AF6_SPI2_USART3_USART4_I2C1,
	}
	I2C0 = I2C2

	// FDCAN1 on PD0 (RX) / PD1 (TX) with onboard transceiver
	CAN1  = &_CAN1
	_CAN1 = FDCAN{
		Bus:             stm32.FDCAN1,
		TxAltFuncSelect: AF3_FDCAN1_FDCAN2,
		RxAltFuncSelect: AF3_FDCAN1_FDCAN2,
		instance:        0,
	}
	// Alias for convenience
	CAN0 = CAN1
)

// Suppress unused import warning for interrupt package
var _ = interrupt.New

func init() {
	// No UART configured on this board - uses USB or CAN for communication
}
