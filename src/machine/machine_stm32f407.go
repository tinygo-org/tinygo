// +build stm32,stm32f407

package machine

// Peripheral abstraction layer for the stm32.

import (
	"device/stm32"
	"runtime/interrupt"
	"unsafe"
)

func CPUFrequency() uint32 {
	return 168000000
}

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
	case 5:
		return stm32.GPIOF
	case 6:
		return stm32.GPIOG
	case 7:
		return stm32.GPIOH
	case 8:
		return stm32.GPIOI
	default:
		panic("machine: unknown port")
	}
}

// enableClock enables the clock for this desired GPIO port.
func (p Pin) enableClock() {
	switch p / 16 {
	case 0:
		stm32.RCC.AHB1ENR.SetBits(stm32.RCC_AHB1ENR_GPIOAEN)
	case 1:
		stm32.RCC.AHB1ENR.SetBits(stm32.RCC_AHB1ENR_GPIOBEN)
	case 2:
		stm32.RCC.AHB1ENR.SetBits(stm32.RCC_AHB1ENR_GPIOCEN)
	case 3:
		stm32.RCC.AHB1ENR.SetBits(stm32.RCC_AHB1ENR_GPIODEN)
	case 4:
		stm32.RCC.AHB1ENR.SetBits(stm32.RCC_AHB1ENR_GPIOEEN)
	case 5:
		stm32.RCC.AHB1ENR.SetBits(stm32.RCC_AHB1ENR_GPIOFEN)
	case 6:
		stm32.RCC.AHB1ENR.SetBits(stm32.RCC_AHB1ENR_GPIOGEN)
	case 7:
		stm32.RCC.AHB1ENR.SetBits(stm32.RCC_AHB1ENR_GPIOHEN)
	case 8:
		stm32.RCC.AHB1ENR.SetBits(stm32.RCC_AHB1ENR_GPIOIEN)
	default:
		panic("machine: unknown port")
	}
}

// Enable peripheral clock
func enableAltFuncClock(bus unsafe.Pointer) {
	switch bus {
	case unsafe.Pointer(stm32.USART1):
		stm32.RCC.APB2ENR.SetBits(stm32.RCC_APB2ENR_USART1EN)
	case unsafe.Pointer(stm32.USART2):
		stm32.RCC.APB1ENR.SetBits(stm32.RCC_APB1ENR_USART2EN)
	case unsafe.Pointer(stm32.I2C1):
		stm32.RCC.APB1ENR.SetBits(stm32.RCC_APB1ENR_I2C1EN)
	case unsafe.Pointer(stm32.SPI1):
		stm32.RCC.APB2ENR.SetBits(stm32.RCC_APB2ENR_SPI1EN)
	}
}

// UART
type UART struct {
	Buffer          *RingBuffer
	Bus             *stm32.USART_Type
	AltFuncSelector stm32.AltFunc
}

var (
	UART0 = UART{
		Buffer:          NewRingBuffer(),
		Bus:             stm32.USART2,
		AltFuncSelector: stm32.AF7_USART1_2_3,
	}
	UART1 = &UART0
)

// Configure the UART.
func (uart UART) Configure(config UARTConfig) {
	// Default baud rate to 115200.
	if config.BaudRate == 0 {
		config.BaudRate = 115200
	}

	// Set the GPIO pins to defaults if they're not set
	if config.TX == 0 && config.RX == 0 {
		config.TX = UART_TX_PIN
		config.RX = UART_RX_PIN
	}

	// Enable USART clock
	enableAltFuncClock(unsafe.Pointer(uart.Bus))

	// use standard TX/RX pins PA2 and PA3
	config.TX.ConfigureAltFunc(PinConfig{Mode: PinModeUARTTX}, uart.AltFuncSelector)
	config.RX.ConfigureAltFunc(PinConfig{Mode: PinModeUARTRX}, uart.AltFuncSelector)

	/*
	  Set baud rate(115200)
	  OVER8 = 0, APB1 = 42mhz
	  +----------+--------+
	  | baudrate | BRR    |
	  +----------+--------+
	  | 1200     | 0x88B8 |
	  | 2400     | 0x445C |
	  | 9600     | 0x1117 |
	  | 19200    | 0x88C  |
	  | 38400    | 0x446  |
	  | 57600    | 0x2D9  |
	  | 115200   | 0x16D  |
	  +----------+--------+
	*/
	uart.Bus.BRR.Set(0x16c)

	// Enable USART2 port.
	uart.Bus.CR1.Set(stm32.USART_CR1_TE | stm32.USART_CR1_RE | stm32.USART_CR1_RXNEIE | stm32.USART_CR1_UE)

	// Enable RX IRQ. TODO: pick the right IRQ_xxx for the bus and the uart
	intr := interrupt.New(stm32.IRQ_USART2, func(interrupt.Interrupt) {
		UART1.Receive(byte((UART1.Bus.DR.Get() & 0xFF)))
	})
	intr.SetPriority(0xc0)
	intr.Enable()
}

// WriteByte writes a byte of data to the UART.
func (uart UART) WriteByte(c byte) error {
	uart.Bus.DR.Set(uint32(c))

	for !uart.Bus.SR.HasBits(stm32.USART_SR_TXE) {
	}
	return nil
}
