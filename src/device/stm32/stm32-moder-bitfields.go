// Hand created file. DO NOT DELETE.
// STM32FXXX (except stm32f1xx) bitfield definitions that are not
//  auto-generated by gen-device-svd.go

// These apply to the stm32 families that use the MODER, OTYPE amd PUPDR
//  registers for managing GPIO functionality.

// Add in other families that use the same settings, e.g. stm32f0xx, etc

// +build stm32,!stm32f103xx

package stm32

// AltFunc represents the alternate function peripherals that can be mapped to
// the GPIO ports. Since these differ by what is supported on the stm32 family
// they are defined in the more specific files
type AltFunc uint8

// Family-wide common post-reset AltFunc. This represents
//  normal GPIO operation of the pins
const AF0_SYSTEM AltFunc = 0

const (
	// Register values for the chip
	// GPIOx_MODER
	GPIOModeInput         = 0
	GPIOModeOutputGeneral = 1
	GPIOModeOutputAltFunc = 2
	GPIOModeAnalog        = 3

	// GPIOx_OTYPER
	GPIOOutputTypePushPull  = 0
	GPIOOutputTypeOpenDrain = 1

	// GPIOx_OSPEEDR
	GPIOSpeedLow      = 0
	GPIOSpeedMid      = 1
	GPIOSpeedHigh     = 2 // Note: this is also low speed on stm32f0, see RM0091
	GPIOSpeedVeryHigh = 3

	// GPIOx_PUPDR
	GPIOPUPDRFloating = 0
	GPIOPUPDRPullUp   = 1
	GPIOPUPDRPullDown = 2
)

// SPI prescaler values fPCLK / X
const (
	SPI_PCLK_2   = 0
	SPI_PCLK_4   = 1
	SPI_PCLK_8   = 2
	SPI_PCLK_16  = 3
	SPI_PCLK_32  = 4
	SPI_PCLK_64  = 5
	SPI_PCLK_128 = 6
	SPI_PCLK_256 = 7
)
