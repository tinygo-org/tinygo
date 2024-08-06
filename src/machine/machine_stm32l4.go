//go:build stm32l4

package machine

import (
	"device/stm32"
	"errors"
	"internal/binary"
	"runtime/interrupt"
	"runtime/volatile"
	"unsafe"
)

// Peripheral abstraction layer for the stm32l4

var deviceIDAddr = []uintptr{0x1FFF7590, 0x1FFF7594, 0x1FFF7598}

const (
	AF0_SYSTEM             = 0
	AF1_TIM1_2_LPTIM1      = 1
	AF2_TIM1_2             = 2
	AF3_USART2             = 3
	AF4_I2C1_2_3           = 4
	AF5_SPI1_2             = 5
	AF6_SPI3               = 6
	AF7_USART1_2_3         = 7
	AF8_LPUART1            = 8
	AF9_CAN1_TSC           = 9
	AF10_USB_QUADSPI       = 10
	AF12_COMP1_2_SWPMI1    = 12
	AF13_SAI1              = 13
	AF14_TIM2_15_16_LPTIM2 = 14
	AF15_EVENTOUT          = 15
)

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
)

// IRQs are defined here as they vary in the SVDs, but do have consistent mapping
// to Timer Interrupts.
const (
	irq_TIM1_BRK_TIM15     = 24
	irq_TIM1_UP_TIM16      = 25
	irq_TIM1_TRG_COM_TIM17 = 26
	irq_TIM1_CC            = 27
	irq_TIM2               = 28
	irq_TIM3               = 29
	irq_TIM4               = 30
	irq_TIM5               = 50
	irq_TIM6               = 54
	irq_TIM7               = 55
	irq_TIM8_BRK           = 43
	irq_TIM8_UP            = 44
	irq_TIM8_TRG_COM       = 45
	irq_TIM8_CC            = 46
)

func (p Pin) getPort() *stm32.GPIO_Type {
	switch p / 16 {
	case 0:
		return stm32.GPIOA
	case 1:
		return stm32.GPIOB
	case 2:
		return stm32.GPIOC
	case 3:
		return stm32.GPIOD
	case 4:
		return stm32.GPIOE
	default:
		panic("machine: unknown port")
	}
}

// enableClock enables the clock for this desired GPIO port.
func (p Pin) enableClock() {
	switch p / 16 {
	case 0:
		stm32.RCC.AHB2ENR.SetBits(stm32.RCC_AHB2ENR_GPIOAEN)
	case 1:
		stm32.RCC.AHB2ENR.SetBits(stm32.RCC_AHB2ENR_GPIOBEN)
	case 2:
		stm32.RCC.AHB2ENR.SetBits(stm32.RCC_AHB2ENR_GPIOCEN)
	case 3:
		stm32.RCC.AHB2ENR.SetBits(stm32.RCC_AHB2ENR_GPIODEN)
	case 4:
		stm32.RCC.AHB2ENR.SetBits(stm32.RCC_AHB2ENR_GPIOEEN)
	default:
		panic("machine: unknown port")
	}
}

// Enable peripheral clock
func enableAltFuncClock(bus unsafe.Pointer) {
	switch bus {
	case unsafe.Pointer(stm32.PWR): // Power interface clock enable
		stm32.RCC.APB1ENR1.SetBits(stm32.RCC_APB1ENR1_PWREN)
	case unsafe.Pointer(stm32.I2C3): // I2C3 clock enable
		stm32.RCC.APB1ENR1.SetBits(stm32.RCC_APB1ENR1_I2C3EN)
	case unsafe.Pointer(stm32.I2C2): // I2C2 clock enable
		stm32.RCC.APB1ENR1.SetBits(stm32.RCC_APB1ENR1_I2C2EN)
	case unsafe.Pointer(stm32.I2C1): // I2C1 clock enable
		stm32.RCC.APB1ENR1.SetBits(stm32.RCC_APB1ENR1_I2C1EN)
	case unsafe.Pointer(stm32.UART4): // UART4 clock enable
		stm32.RCC.APB1ENR1.SetBits(stm32.RCC_APB1ENR1_UART4EN)
	case unsafe.Pointer(stm32.USART3): // USART3 clock enable
		stm32.RCC.APB1ENR1.SetBits(stm32.RCC_APB1ENR1_USART3EN)
	case unsafe.Pointer(stm32.USART2): // USART2 clock enable
		stm32.RCC.APB1ENR1.SetBits(stm32.RCC_APB1ENR1_USART2EN)
	case unsafe.Pointer(stm32.SPI3): // SPI3 clock enable
		stm32.RCC.APB1ENR1.SetBits(stm32.RCC_APB1ENR1_SPI3EN)
	case unsafe.Pointer(stm32.SPI2): // SPI2 clock enable
		stm32.RCC.APB1ENR1.SetBits(stm32.RCC_APB1ENR1_SPI2EN)
	case unsafe.Pointer(stm32.WWDG): // Window watchdog clock enable
		stm32.RCC.APB1ENR1.SetBits(stm32.RCC_APB1ENR1_WWDGEN)
	case unsafe.Pointer(stm32.TIM7): // TIM7 clock enable
		stm32.RCC.APB1ENR1.SetBits(stm32.RCC_APB1ENR1_TIM7EN)
	case unsafe.Pointer(stm32.TIM6): // TIM6 clock enable
		stm32.RCC.APB1ENR1.SetBits(stm32.RCC_APB1ENR1_TIM6EN)
	case unsafe.Pointer(stm32.TIM3): // TIM3 clock enable
		stm32.RCC.APB1ENR1.SetBits(stm32.RCC_APB1ENR1_TIM3EN)
	case unsafe.Pointer(stm32.TIM2): // TIM2 clock enable
		stm32.RCC.APB1ENR1.SetBits(stm32.RCC_APB1ENR1_TIM2EN)
	case unsafe.Pointer(stm32.LPTIM2): // LPTIM2 clock enable
		stm32.RCC.APB1ENR2.SetBits(stm32.RCC_APB1ENR2_LPTIM2EN)
	case unsafe.Pointer(stm32.LPUART1): // LPUART1 clock enable
		stm32.RCC.APB1ENR2.SetBits(stm32.RCC_APB1ENR2_LPUART1EN)
	case unsafe.Pointer(stm32.TIM16): // TIM16 clock enable
		stm32.RCC.APB2ENR.SetBits(stm32.RCC_APB2ENR_TIM16EN)
	case unsafe.Pointer(stm32.TIM15): // TIM15 clock enable
		stm32.RCC.APB2ENR.SetBits(stm32.RCC_APB2ENR_TIM15EN)
	case unsafe.Pointer(stm32.SYSCFG): // System configuration controller clock enable
		stm32.RCC.APB2ENR.SetBits(stm32.RCC_APB2ENR_SYSCFGEN)
	case unsafe.Pointer(stm32.SPI1): // SPI1 clock enable
		stm32.RCC.APB2ENR.SetBits(stm32.RCC_APB2ENR_SPI1EN)
	case unsafe.Pointer(stm32.USART1): // USART1 clock enable
		stm32.RCC.APB2ENR.SetBits(stm32.RCC_APB2ENR_USART1EN)
	case unsafe.Pointer(stm32.TIM1): // TIM1 clock enable
		stm32.RCC.APB2ENR.SetBits(stm32.RCC_APB2ENR_TIM1EN)
	}
}

func handlePinInterrupt(pin uint8) {
	if stm32.EXTI.PR1.HasBits(1 << pin) {
		// Writing 1 to the pending register clears the
		// pending flag for that bit
		stm32.EXTI.PR1.Set(1 << pin)

		callback := pinCallbacks[pin]
		if callback != nil {
			callback(interruptPins[pin])
		}
	}
}

func (p Pin) registerInterrupt() interrupt.Interrupt {
	pin := uint8(p) % 16

	switch pin {
	case 0:
		return interrupt.New(stm32.IRQ_EXTI0, func(interrupt.Interrupt) { handlePinInterrupt(0) })
	case 1:
		return interrupt.New(stm32.IRQ_EXTI1, func(interrupt.Interrupt) { handlePinInterrupt(1) })
	case 2:
		return interrupt.New(stm32.IRQ_EXTI2, func(interrupt.Interrupt) { handlePinInterrupt(2) })
	case 3:
		return interrupt.New(stm32.IRQ_EXTI3, func(interrupt.Interrupt) { handlePinInterrupt(3) })
	case 4:
		return interrupt.New(stm32.IRQ_EXTI4, func(interrupt.Interrupt) { handlePinInterrupt(4) })
	case 5:
		return interrupt.New(stm32.IRQ_EXTI9_5, func(interrupt.Interrupt) { handlePinInterrupt(5) })
	case 6:
		return interrupt.New(stm32.IRQ_EXTI9_5, func(interrupt.Interrupt) { handlePinInterrupt(6) })
	case 7:
		return interrupt.New(stm32.IRQ_EXTI9_5, func(interrupt.Interrupt) { handlePinInterrupt(7) })
	case 8:
		return interrupt.New(stm32.IRQ_EXTI9_5, func(interrupt.Interrupt) { handlePinInterrupt(8) })
	case 9:
		return interrupt.New(stm32.IRQ_EXTI9_5, func(interrupt.Interrupt) { handlePinInterrupt(9) })
	case 10:
		return interrupt.New(stm32.IRQ_EXTI15_10, func(interrupt.Interrupt) { handlePinInterrupt(10) })
	case 11:
		return interrupt.New(stm32.IRQ_EXTI15_10, func(interrupt.Interrupt) { handlePinInterrupt(11) })
	case 12:
		return interrupt.New(stm32.IRQ_EXTI15_10, func(interrupt.Interrupt) { handlePinInterrupt(12) })
	case 13:
		return interrupt.New(stm32.IRQ_EXTI15_10, func(interrupt.Interrupt) { handlePinInterrupt(13) })
	case 14:
		return interrupt.New(stm32.IRQ_EXTI15_10, func(interrupt.Interrupt) { handlePinInterrupt(14) })
	case 15:
		return interrupt.New(stm32.IRQ_EXTI15_10, func(interrupt.Interrupt) { handlePinInterrupt(15) })
	}

	return interrupt.Interrupt{}
}

//---------- UART related code

// Configure the UART.
func (uart *UART) configurePins(config UARTConfig) {
	// enable the alternate functions on the TX and RX pins
	config.TX.ConfigureAltFunc(PinConfig{Mode: PinModeUARTTX}, uart.TxAltFuncSelector)
	config.RX.ConfigureAltFunc(PinConfig{Mode: PinModeUARTRX}, uart.RxAltFuncSelector)
}

// UART baudrate calc based on the bus and clockspeed
// NOTE: keep this in sync with the runtime/runtime_stm32l5x2.go clock init code
func (uart *UART) getBaudRateDivisor(baudRate uint32) uint32 {
	return (CPUFrequency() / baudRate)
}

// Register names vary by ST processor, these are for STM L5
func (uart *UART) setRegisters() {
	uart.rxReg = &uart.Bus.RDR
	uart.txReg = &uart.Bus.TDR
	uart.statusReg = &uart.Bus.ISR
	uart.txEmptyFlag = stm32.USART_ISR_TXE
}

//---------- SPI related types and code

// SPI on the STM32Fxxx using MODER / alternate function pins
type SPI struct {
	Bus             *stm32.SPI_Type
	AltFuncSelector uint8
}

func (spi SPI) config8Bits() {
	// Set rx threshold to 8-bits, so RXNE flag is set for 1 byte
	// (common STM32 SPI implementation does 8-bit transfers only)
	spi.Bus.CR2.SetBits(stm32.SPI_CR2_FRXTH)
}

// Set baud rate for SPI
func (spi SPI) getBaudRate(config SPIConfig) uint32 {
	var conf uint32

	// Default
	if config.Frequency == 0 {
		config.Frequency = 4e6
	}

	localFrequency := config.Frequency

	// set frequency dependent on PCLK prescaler. Since these are rather weird
	// speeds due to the CPU frequency, pick a range up to that frequency for
	// clients to use more human-understandable numbers, e.g. nearest 100KHz

	// These are based on 80MHz peripheral clock frequency
	switch {
	case localFrequency < 312500:
		conf = stm32.SPI_CR1_BR_Div256
	case localFrequency < 625000:
		conf = stm32.SPI_CR1_BR_Div128
	case localFrequency < 1250000:
		conf = stm32.SPI_CR1_BR_Div64
	case localFrequency < 2500000:
		conf = stm32.SPI_CR1_BR_Div32
	case localFrequency < 5000000:
		conf = stm32.SPI_CR1_BR_Div16
	case localFrequency < 10000000:
		conf = stm32.SPI_CR1_BR_Div8
		// NOTE: many SPI components won't operate reliably (or at all) above 10MHz
		// Check the datasheet of the part
	case localFrequency < 20000000:
		conf = stm32.SPI_CR1_BR_Div4
	case localFrequency < 40000000:
		conf = stm32.SPI_CR1_BR_Div2
	default:
		// None of the specific baudrates were selected; choose the lowest speed
		conf = stm32.SPI_CR1_BR_Div256
	}

	return conf << stm32.SPI_CR1_BR_Pos
}

// Configure SPI pins for input output and clock
func (spi SPI) configurePins(config SPIConfig) {
	config.SCK.ConfigureAltFunc(PinConfig{Mode: PinModeSPICLK}, spi.AltFuncSelector)
	config.SDO.ConfigureAltFunc(PinConfig{Mode: PinModeSPISDO}, spi.AltFuncSelector)
	config.SDI.ConfigureAltFunc(PinConfig{Mode: PinModeSPISDI}, spi.AltFuncSelector)
}

//---------- Timer related code

var (
	TIM1 = TIM{
		EnableRegister: &stm32.RCC.APB2ENR,
		EnableFlag:     stm32.RCC_APB2ENR_TIM1EN,
		Device:         stm32.TIM1,
		Channels: [4]TimerChannel{
			TimerChannel{Pins: []PinFunction{
				{PA8, AF1_TIM1_2_LPTIM1},
			}},
			TimerChannel{Pins: []PinFunction{
				{PA9, AF1_TIM1_2_LPTIM1},
			}},
			TimerChannel{Pins: []PinFunction{
				{PA10, AF1_TIM1_2_LPTIM1},
			}},
			TimerChannel{Pins: []PinFunction{
				{PA11, AF1_TIM1_2_LPTIM1},
			}},
		},
		busFreq: APB2_TIM_FREQ,
	}

	TIM2 = TIM{
		EnableRegister: &stm32.RCC.APB1ENR1,
		EnableFlag:     stm32.RCC_APB1ENR1_TIM2EN,
		Device:         stm32.TIM2,
		Channels: [4]TimerChannel{
			TimerChannel{Pins: []PinFunction{
				{PA0, AF1_TIM1_2_LPTIM1},
				{PA5, AF1_TIM1_2_LPTIM1},
				{PA15, AF1_TIM1_2_LPTIM1},
			}},
			TimerChannel{Pins: []PinFunction{
				{PA1, AF1_TIM1_2_LPTIM1},
				{PB3, AF1_TIM1_2_LPTIM1},
			}},
			TimerChannel{Pins: []PinFunction{
				{PA2, AF1_TIM1_2_LPTIM1},
			}},
			TimerChannel{Pins: []PinFunction{
				{PA3, AF1_TIM1_2_LPTIM1},
			}},
		},
		busFreq: APB1_TIM_FREQ,
	}

	TIM3 = TIM{
		EnableRegister: &stm32.RCC.APB1ENR1,
		EnableFlag:     stm32.RCC_APB1ENR1_TIM3EN,
		Device:         stm32.TIM3,
		Channels: [4]TimerChannel{
			TimerChannel{Pins: []PinFunction{}},
			TimerChannel{Pins: []PinFunction{}},
			TimerChannel{Pins: []PinFunction{}},
			TimerChannel{Pins: []PinFunction{}},
		},
		busFreq: APB1_TIM_FREQ,
	}

	TIM6 = TIM{
		EnableRegister: &stm32.RCC.APB1ENR1,
		EnableFlag:     stm32.RCC_APB1ENR1_TIM6EN,
		Device:         stm32.TIM6,
		Channels: [4]TimerChannel{
			TimerChannel{Pins: []PinFunction{}},
			TimerChannel{Pins: []PinFunction{}},
			TimerChannel{Pins: []PinFunction{}},
			TimerChannel{Pins: []PinFunction{}},
		},
		busFreq: APB1_TIM_FREQ,
	}

	TIM7 = TIM{
		EnableRegister: &stm32.RCC.APB1ENR1,
		EnableFlag:     stm32.RCC_APB1ENR1_TIM7EN,
		Device:         stm32.TIM7,
		Channels: [4]TimerChannel{
			TimerChannel{Pins: []PinFunction{}},
			TimerChannel{Pins: []PinFunction{}},
			TimerChannel{Pins: []PinFunction{}},
			TimerChannel{Pins: []PinFunction{}},
		},
		busFreq: APB1_TIM_FREQ,
	}

	TIM15 = TIM{
		EnableRegister: &stm32.RCC.APB2ENR,
		EnableFlag:     stm32.RCC_APB2ENR_TIM15EN,
		Device:         stm32.TIM15,
		Channels: [4]TimerChannel{
			TimerChannel{Pins: []PinFunction{
				{PA2, AF14_TIM2_15_16_LPTIM2},
			}},
			TimerChannel{Pins: []PinFunction{
				{PA3, AF14_TIM2_15_16_LPTIM2},
			}},
			TimerChannel{Pins: []PinFunction{}},
			TimerChannel{Pins: []PinFunction{}},
		},
		busFreq: APB2_TIM_FREQ,
	}

	TIM16 = TIM{
		EnableRegister: &stm32.RCC.APB2ENR,
		EnableFlag:     stm32.RCC_APB2ENR_TIM16EN,
		Device:         stm32.TIM16,
		Channels: [4]TimerChannel{
			TimerChannel{Pins: []PinFunction{
				{PA6, AF14_TIM2_15_16_LPTIM2},
			}},
			TimerChannel{Pins: []PinFunction{}},
			TimerChannel{Pins: []PinFunction{}},
			TimerChannel{Pins: []PinFunction{}},
		},
		busFreq: APB2_TIM_FREQ,
	}
)

func (t *TIM) registerUPInterrupt() interrupt.Interrupt {
	switch t {
	case &TIM1:
		return interrupt.New(irq_TIM1_UP_TIM16, TIM1.handleUPInterrupt)
	case &TIM2:
		return interrupt.New(irq_TIM2, TIM2.handleUPInterrupt)
	case &TIM3:
		return interrupt.New(irq_TIM3, TIM3.handleUPInterrupt)
	case &TIM6:
		return interrupt.New(irq_TIM6, TIM6.handleUPInterrupt)
	case &TIM7:
		return interrupt.New(irq_TIM7, TIM7.handleUPInterrupt)
	case &TIM15:
		return interrupt.New(irq_TIM1_BRK_TIM15, TIM15.handleUPInterrupt)
	case &TIM16:
		return interrupt.New(irq_TIM1_UP_TIM16, TIM16.handleUPInterrupt)
	}

	return interrupt.Interrupt{}
}

func (t *TIM) registerOCInterrupt() interrupt.Interrupt {
	switch t {
	case &TIM1:
		return interrupt.New(irq_TIM1_CC, TIM1.handleUPInterrupt)
	case &TIM2:
		return interrupt.New(irq_TIM2, TIM2.handleOCInterrupt)
	case &TIM3:
		return interrupt.New(irq_TIM3, TIM3.handleOCInterrupt)
	case &TIM6:
		return interrupt.New(irq_TIM6, TIM6.handleOCInterrupt)
	case &TIM7:
		return interrupt.New(irq_TIM7, TIM7.handleOCInterrupt)
	case &TIM15:
		return interrupt.New(irq_TIM1_BRK_TIM15, TIM15.handleOCInterrupt)
	case &TIM16:
		return interrupt.New(irq_TIM1_UP_TIM16, TIM16.handleOCInterrupt)
	}

	return interrupt.Interrupt{}
}

func (t *TIM) enableMainOutput() {
	// nothing to do - no BDTR register
}

type arrtype = uint32
type arrRegType = volatile.Register32

const (
	ARR_MAX = 0x10000
	PSC_MAX = 0x10000
)

func initRNG() {
	stm32.RCC.CRRCR.SetBits(stm32.RCC_CRRCR_HSI48ON)
	for !stm32.RCC.CRRCR.HasBits(stm32.RCC_CRRCR_HSI48RDY) {
	}

	stm32.RCC.AHB2ENR.SetBits(stm32.RCC_AHB2ENR_RNGEN)
	stm32.RNG.CR.SetBits(stm32.RNG_CR_RNGEN)
}

//---------- Flash related code

const eraseBlockSizeValue = 2048

// see RM0394 page 83
// eraseBlock of the passed in block number
func eraseBlock(block uint32) error {
	waitUntilFlashDone()

	// clear any previous errors
	stm32.FLASH.SR.SetBits(0x3FA)

	// page erase operation
	stm32.FLASH.SetCR_PER(1)
	defer stm32.FLASH.SetCR_PER(0)

	// set the page to be erased
	stm32.FLASH.SetCR_PNB(block)

	// start the page erase
	stm32.FLASH.SetCR_START(1)

	waitUntilFlashDone()

	if err := checkError(); err != nil {
		return err
	}

	return nil
}

const writeBlockSize = 8

// see RM0394 page 84
// It is only possible to program double word (2 x 32-bit data).
func writeFlashData(address uintptr, data []byte) (int, error) {
	if len(data)%writeBlockSize != 0 {
		return 0, errFlashInvalidWriteLength
	}

	waitUntilFlashDone()

	// clear any previous errors
	stm32.FLASH.SR.SetBits(0x3FA)

	for j := 0; j < len(data); j += writeBlockSize {
		// start page write operation
		stm32.FLASH.SetCR_PG(1)

		// write second word using double-word high order word
		*(*uint32)(unsafe.Pointer(address)) = binary.LittleEndian.Uint32(data[j : j+writeBlockSize/2])

		address += writeBlockSize / 2

		// write first word using double-word low order word
		*(*uint32)(unsafe.Pointer(address)) = binary.LittleEndian.Uint32(data[j+writeBlockSize/2 : j+writeBlockSize])

		waitUntilFlashDone()

		if err := checkError(); err != nil {
			return j, err
		}

		// end flash write
		stm32.FLASH.SetCR_PG(0)
		address += writeBlockSize / 2
	}

	return len(data), nil
}

func waitUntilFlashDone() {
	for stm32.FLASH.GetSR_BSY() != 0 {
	}
}

var (
	errFlashPGS  = errors.New("errFlashPGS")
	errFlashSIZE = errors.New("errFlashSIZE")
	errFlashPGA  = errors.New("errFlashPGA")
	errFlashWRP  = errors.New("errFlashWRP")
	errFlashPROG = errors.New("errFlashPROG")
)

func checkError() error {
	switch {
	case stm32.FLASH.GetSR_PGSERR() != 0:
		return errFlashPGS
	case stm32.FLASH.GetSR_SIZERR() != 0:
		return errFlashSIZE
	case stm32.FLASH.GetSR_PGAERR() != 0:
		return errFlashPGA
	case stm32.FLASH.GetSR_WRPERR() != 0:
		return errFlashWRP
	case stm32.FLASH.GetSR_PROGERR() != 0:
		return errFlashPROG
	}

	return nil
}
