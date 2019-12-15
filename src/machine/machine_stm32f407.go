// +build stm32,stm32f407

package machine

// Peripheral abstraction layer for the stm32.

import (
	"device/stm32"
	"runtime/interrupt"
)

func CPUFrequency() uint32 {
	return 168000000
}

const (
	// Mode Flag
	PinOutput        PinMode = 0
	PinInput         PinMode = PinInputFloating
	PinInputFloating PinMode = 1
	PinInputPulldown PinMode = 2
	PinInputPullup   PinMode = 3

	// for UART
	PinModeUartTX PinMode = 4
	PinModeUartRX PinMode = 5

	//GPIOx_MODER
	GPIO_MODE_INPUT          = 0
	GPIO_MODE_GENERAL_OUTPUT = 1
	GPIO_MODE_ALTERNABTIVE   = 2
	GPIO_MODE_ANALOG         = 3

	//GPIOx_OTYPER
	GPIO_OUTPUT_MODE_PUSH_PULL  = 0
	GPIO_OUTPUT_MODE_OPEN_DRAIN = 1

	// GPIOx_OSPEEDR
	GPIO_SPEED_LOW     = 0
	GPIO_SPEED_MID     = 1
	GPIO_SPEED_HI      = 2
	GPIO_SPEED_VERY_HI = 3

	// GPIOx_PUPDR
	GPIO_FLOATING  = 0
	GPIO_PULL_UP   = 1
	GPIO_PULL_DOWN = 2
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

// Configure this pin with the given configuration.
func (p Pin) Configure(config PinConfig) {
	// Configure the GPIO pin.
	p.enableClock()
	port := p.getPort()
	pin := uint8(p) % 16
	pos := pin * 2

	if config.Mode == PinInputFloating {
		port.MODER.Set((uint32(port.MODER.Get())&^(0x3<<pos) | (uint32(GPIO_MODE_INPUT) << pos)))
		port.PUPDR.Set((uint32(port.PUPDR.Get())&^(0x3<<pos) | (uint32(GPIO_FLOATING) << pos)))
	} else if config.Mode == PinInputPulldown {
		port.MODER.Set((uint32(port.MODER.Get())&^(0x3<<pos) | (uint32(GPIO_MODE_INPUT) << pos)))
		port.PUPDR.Set((uint32(port.PUPDR.Get())&^(0x3<<pos) | (uint32(GPIO_PULL_DOWN) << pos)))
	} else if config.Mode == PinInputPullup {
		port.MODER.Set((uint32(port.MODER.Get())&^(0x3<<pos) | (uint32(GPIO_MODE_INPUT) << pos)))
		port.PUPDR.Set((uint32(port.PUPDR.Get())&^(0x3<<pos) | (uint32(GPIO_PULL_UP) << pos)))
	} else if config.Mode == PinOutput {
		port.MODER.Set((uint32(port.MODER.Get())&^(0x3<<pos) | (uint32(GPIO_MODE_GENERAL_OUTPUT) << pos)))
		port.OSPEEDR.Set((uint32(port.OSPEEDR.Get())&^(0x3<<pos) | (uint32(GPIO_SPEED_HI) << pos)))
	} else if config.Mode == PinModeUartTX {
		port.MODER.Set((uint32(port.MODER.Get())&^(0x3<<pos) | (uint32(GPIO_MODE_ALTERNABTIVE) << pos)))
		port.OSPEEDR.Set((uint32(port.OSPEEDR.Get())&^(0x3<<pos) | (uint32(GPIO_SPEED_HI) << pos)))
		port.PUPDR.Set((uint32(port.PUPDR.Get())&^(0x3<<pos) | (uint32(GPIO_PULL_UP) << pos)))
		p.setAltFunc(0x7)
	} else if config.Mode == PinModeUartRX {
		port.MODER.Set((uint32(port.MODER.Get())&^(0x3<<pos) | (uint32(GPIO_MODE_ALTERNABTIVE) << pos)))
		port.PUPDR.Set((uint32(port.PUPDR.Get())&^(0x3<<pos) | (uint32(GPIO_FLOATING) << pos)))
		p.setAltFunc(0x7)
	}
}

func (p Pin) setAltFunc(af uint32) {
	port := p.getPort()
	pin := uint8(p) % 16
	pos := pin * 4
	if pin >= 8 {
		port.AFRH.Set(uint32(port.AFRH.Get())&^(0xF<<pos) | ((af & 0xF) << pos))
	} else {
		port.AFRL.Set(uint32(port.AFRL.Get())&^(0xF<<pos) | ((af & 0xF) << pos))
	}
}

// Set the pin to high or low.
// Warning: only use this on an output pin!
func (p Pin) Set(high bool) {
	port := p.getPort()
	pin := p % 16
	if high {
		port.BSRR.Set(1 << uint8(pin))
	} else {
		port.BSRR.Set(1 << uint8(pin+16))
	}
}

// UART
type UART struct {
	Buffer *RingBuffer
}

var (
	// Both UART0 and UART1 refer to USART2.
	UART0 = UART{Buffer: NewRingBuffer()}
	UART1 = &UART0
)

// Configure the UART.
func (uart UART) Configure(config UARTConfig) {
	// Default baud rate to 115200.
	if config.BaudRate == 0 {
		config.BaudRate = 115200
	}

	// pins
	switch config.TX {
	default:
		// use standard TX/RX pins PA2 and PA3
		UART_TX_PIN.Configure(PinConfig{Mode: PinModeUartTX})
		UART_RX_PIN.Configure(PinConfig{Mode: PinModeUartRX})
	}

	// Enable USART2 clock
	stm32.RCC.APB1ENR.SetBits(stm32.RCC_APB1ENR_USART2EN)

	/*
	  Set baud rate(115200)
	  OVER8 = 0, APB2 = 42mhz
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
	stm32.USART2.BRR.Set(0x16c)

	// Enable USART2 port.
	stm32.USART2.CR1.Set(stm32.USART_CR1_TE | stm32.USART_CR1_RE | stm32.USART_CR1_RXNEIE | stm32.USART_CR1_UE)

	// Enable RX IRQ.
	intr := interrupt.New(stm32.IRQ_USART2, func(interrupt.Interrupt) {
		UART1.Receive(byte((stm32.USART2.DR.Get() & 0xFF)))
	})
	intr.SetPriority(0xc0)
	intr.Enable()
}

// WriteByte writes a byte of data to the UART.
func (uart UART) WriteByte(c byte) error {
	stm32.USART2.DR.Set(uint32(c))

	for !stm32.USART2.SR.HasBits(stm32.USART_SR_TXE) {
	}
	return nil
}
