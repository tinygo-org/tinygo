// +build esp32

package machine

import "device/esp"

const peripheralClock = 80000000 // 80MHz

type PinMode uint8

const (
	PinOutput PinMode = iota
	PinInput
)

// Configure this pin with the given configuration.
func (p Pin) Configure(config PinConfig) {
	if config.Mode == PinOutput {
		// Set the 'output enable' bit.
		if p < 32 {
			esp.GPIO.ENABLE_W1TS.Set(1 << p)
		} else {
			esp.GPIO.ENABLE1_W1TS.Set(1 << (p - 32))
		}
	} else {
		// Clear the 'output enable' bit.
		if p < 32 {
			esp.GPIO.ENABLE_W1TC.Set(1 << p)
		} else {
			esp.GPIO.ENABLE1_W1TC.Set(1 << (p - 32))
		}
	}
}

// Set the pin to high or low.
// Warning: only use this on an output pin!
func (p Pin) Set(value bool) {
	if value {
		if p < 32 {
			esp.GPIO.OUT_W1TS.Set(1 << p)
		} else {
			esp.GPIO.OUT1_W1TS.Set(1 << (p - 32))
		}
	} else {
		if p < 32 {
			esp.GPIO.OUT_W1TC.Set(1 << p)
		} else {
			esp.GPIO.OUT1_W1TC.Set(1 << (p - 32))
		}
	}
}

// Get returns the current value of a GPIO pin when the pin is configured as an
// input.
func (p Pin) Get() bool {
	if p < 32 {
		return esp.GPIO.IN.Get()&(1<<p) != 0
	} else {
		return esp.GPIO.IN1.Get()&(1<<(p-32)) != 0
	}
}

var (
	UART0 = UART{Bus: esp.UART0, Buffer: NewRingBuffer()}
	UART1 = UART{Bus: esp.UART1, Buffer: NewRingBuffer()}
	UART2 = UART{Bus: esp.UART2, Buffer: NewRingBuffer()}
)

type UART struct {
	Bus    *esp.UART_Type
	Buffer *RingBuffer
}

func (uart UART) Configure(config UARTConfig) {
	if config.BaudRate == 0 {
		config.BaudRate = 115200
	}
	uart.Bus.CLKDIV.Set(peripheralClock / config.BaudRate)
}

func (uart UART) WriteByte(b byte) error {
	for (uart.Bus.STATUS.Get()>>16)&0xff >= 128 {
		// Read UART_TXFIFO_CNT from the status register, which indicates how
		// many bytes there are in the transmit buffer. Wait until there are
		// less than 128 bytes in this buffer (the default buffer size).
	}
	uart.Bus.TX_FIFO.Set(b)
	return nil
}
