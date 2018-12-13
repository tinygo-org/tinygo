// +build stm32,stm32f103xx

package machine

// Peripheral abstraction layer for the stm32.

import (
	"device/arm"
	"device/stm32"
)

const CPU_FREQUENCY = 72000000

const (
	GPIO_INPUT        = 0 // Input mode
	GPIO_OUTPUT_10MHz = 1 // Output mode, max speed 10MHz
	GPIO_OUTPUT_2MHz  = 2 // Output mode, max speed 2MHz
	GPIO_OUTPUT_50MHz = 3 // Output mode, max speed 50MHz
	GPIO_OUTPUT       = GPIO_OUTPUT_2MHz

	GPIO_INPUT_MODE_ANALOG       = 0  // Input analog mode
	GPIO_INPUT_MODE_FLOATING     = 4  // Input floating mode
	GPIO_INPUT_MODE_PULL_UP_DOWN = 8  // Input pull up/down mode
	GPIO_INPUT_MODE_RESERVED     = 12 // Input mode (reserved)

	GPIO_OUTPUT_MODE_GP_PUSH_PULL   = 0  // Output mode general purpose push/pull
	GPIO_OUTPUT_MODE_GP_OPEN_DRAIN  = 4  // Output mode general purpose open drain
	GPIO_OUTPUT_MODE_ALT_PUSH_PULL  = 8  // Output mode alt. purpose push/pull
	GPIO_OUTPUT_MODE_ALT_OPEN_DRAIN = 12 // Output mode alt. purpose open drain
)

func (p GPIO) getPort() *stm32.GPIO_Type {
	switch p.Pin / 16 {
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
	default:
		panic("machine: unknown port")
	}
}

// enableClock enables the clock for this desired GPIO port.
func (p GPIO) enableClock() {
	switch p.Pin / 16 {
	case 0:
		stm32.RCC.APB2ENR |= stm32.RCC_APB2ENR_IOPAEN
	case 1:
		stm32.RCC.APB2ENR |= stm32.RCC_APB2ENR_IOPBEN
	case 2:
		stm32.RCC.APB2ENR |= stm32.RCC_APB2ENR_IOPCEN
	case 3:
		stm32.RCC.APB2ENR |= stm32.RCC_APB2ENR_IOPDEN
	case 4:
		stm32.RCC.APB2ENR |= stm32.RCC_APB2ENR_IOPEEN
	case 5:
		stm32.RCC.APB2ENR |= stm32.RCC_APB2ENR_IOPFEN
	case 6:
		stm32.RCC.APB2ENR |= stm32.RCC_APB2ENR_IOPGEN
	default:
		panic("machine: unknown port")
	}
}

// Configure this pin with the given configuration.
func (p GPIO) Configure(config GPIOConfig) {
	// Configure the GPIO pin.
	p.enableClock()
	port := p.getPort()
	pin := p.Pin % 16
	pos := p.Pin % 8 * 4
	if pin < 8 {
		port.CRL = stm32.RegValue((uint32(port.CRL) &^ (0xf << pos)) | (uint32(config.Mode) << pos))
	} else {
		port.CRH = stm32.RegValue((uint32(port.CRH) &^ (0xf << pos)) | (uint32(config.Mode) << pos))
	}
}

// Set the pin to high or low.
// Warning: only use this on an output pin!
func (p GPIO) Set(high bool) {
	port := p.getPort()
	pin := p.Pin % 16
	if high {
		port.BSRR = 1 << pin
	} else {
		port.BSRR = 1 << (pin + 16)
	}
}

// UART
var (
	// USART1 is the first hardware serial port on the STM32.
	// Both UART0 and UART1 refers to USART1.
	UART0 = &UART{}
	UART1 = UART0
)

// Configure the UART.
func (uart UART) Configure(config UARTConfig) {
	// Default baud rate to 115200.
	if config.BaudRate == 0 {
		config.BaudRate = 115200
	}

	// pins
	switch config.TX {
	case PB6:
		// use alternate TX/RX pins PB6/PB7 via AFIO mapping
		stm32.RCC.APB2ENR |= stm32.RCC_APB2ENR_AFIOEN
		stm32.AFIO.MAPR |= stm32.AFIO_MAPR_USART1_REMAP
		GPIO{PB6}.Configure(GPIOConfig{Mode: GPIO_OUTPUT_50MHz + GPIO_OUTPUT_MODE_ALT_PUSH_PULL})
		GPIO{PB7}.Configure(GPIOConfig{Mode: GPIO_INPUT_MODE_FLOATING})
	default:
		// use standard TX/RX pins PA9 and PA10
		GPIO{PA9}.Configure(GPIOConfig{Mode: GPIO_OUTPUT_50MHz + GPIO_OUTPUT_MODE_ALT_PUSH_PULL})
		GPIO{PA10}.Configure(GPIOConfig{Mode: GPIO_INPUT_MODE_FLOATING})
	}

	// Enable USART1 clock
	stm32.RCC.APB2ENR |= stm32.RCC_APB2ENR_USART1EN

	// Set baud rate
	uart.SetBaudRate(config.BaudRate)

	// Enable USART1 port.
	stm32.USART1.CR1 = stm32.USART_CR1_TE | stm32.USART_CR1_RE | stm32.USART_CR1_RXNEIE | stm32.USART_CR1_UE

	// Enable RX IRQ.
	arm.EnableIRQ(stm32.IRQ_USART1)
}

// SetBaudRate sets the communication speed for the UART.
func (uart UART) SetBaudRate(br uint32) {
	divider := CPU_FREQUENCY / br
	stm32.USART1.BRR = stm32.RegValue(divider)
}

// WriteByte writes a byte of data to the UART.
func (uart UART) WriteByte(c byte) error {
	stm32.USART1.DR = stm32.RegValue(c)

	for (stm32.USART1.SR & stm32.USART_SR_TXE) == 0 {
	}
	return nil
}

//go:export USART1_IRQHandler
func handleUART1() {
	bufferPut(byte((stm32.USART1.DR & 0xFF)))
}
