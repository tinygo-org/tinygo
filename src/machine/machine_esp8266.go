//go:build esp8266

package machine

import (
	"device/esp"
	"runtime/volatile"
)

const deviceName = esp.Device

func CPUFrequency() uint32 {
	return 80000000 // 80MHz
}

const (
	PinOutput PinMode = iota
	PinInput
)

// Hardware pin numbers
const (
	GPIO0 Pin = iota
	GPIO1
	GPIO2
	GPIO3
	GPIO4
	GPIO5
	GPIO6
	GPIO7
	GPIO8
	GPIO9
	GPIO10
	GPIO11
	GPIO12
	GPIO13
	GPIO14
	GPIO15
	GPIO16
)

// Pins that are fixed by the chip.
const (
	UART_TX_PIN Pin = 1
	UART_RX_PIN Pin = 3
)

// Pin functions are not trivial. The below array maps a pin number (GPIO
// number) to the pad as used in the IO mux.
// Tables with the mapping:
// https://www.esp8266.com/wiki/doku.php?id=esp8266_gpio_pin_allocations#pin_functions
// https://www.espressif.com/sites/default/files/documentation/ESP8266_Pin_List_0.xls
var pinPadMapping = [...]uint8{
	12: 0,
	13: 1,
	14: 2,
	15: 3,
	3:  4,
	1:  5,
	6:  6,
	7:  7,
	8:  8,
	9:  9,
	10: 10,
	11: 11,
	0:  12,
	2:  13,
	4:  14,
	5:  15,
}

// getPad returns the pad number and the register to configure this pad.
func (p Pin) getPad() (uint8, *volatile.Register32) {
	pad := pinPadMapping[p]
	var reg *volatile.Register32
	switch pad {
	case 0:
		reg = &esp.IO_MUX.IO_MUX_MTDI
	case 1:
		reg = &esp.IO_MUX.IO_MUX_MTCK
	case 2:
		reg = &esp.IO_MUX.IO_MUX_MTMS
	case 3:
		reg = &esp.IO_MUX.IO_MUX_MTDO
	case 4:
		reg = &esp.IO_MUX.IO_MUX_U0RXD
	case 5:
		reg = &esp.IO_MUX.IO_MUX_U0TXD
	case 6:
		reg = &esp.IO_MUX.IO_MUX_SD_CLK
	case 7:
		reg = &esp.IO_MUX.IO_MUX_SD_DATA0
	case 8:
		reg = &esp.IO_MUX.IO_MUX_SD_DATA1
	case 9:
		reg = &esp.IO_MUX.IO_MUX_SD_DATA2
	case 10:
		reg = &esp.IO_MUX.IO_MUX_SD_DATA3
	case 11:
		reg = &esp.IO_MUX.IO_MUX_SD_CMD
	case 12:
		reg = &esp.IO_MUX.IO_MUX_GPIO0
	case 13:
		reg = &esp.IO_MUX.IO_MUX_GPIO2
	case 14:
		reg = &esp.IO_MUX.IO_MUX_GPIO4
	case 15:
		reg = &esp.IO_MUX.IO_MUX_GPIO5
	}
	return pad, reg
}

// Configure sets the given pin as output or input pin.
func (p Pin) Configure(config PinConfig) {
	switch config.Mode {
	case PinInput, PinOutput:
		pad, reg := p.getPad()
		if pad >= 12 { // pin 0, 2, 4, 5
			reg.Set(0 << 4) // function 0 at bit position 4
		} else {
			reg.Set(3 << 4) // function 3 at bit position 4
		}
		if config.Mode == PinOutput {
			esp.GPIO.GPIO_ENABLE_W1TS.Set(1 << p)
		} else {
			esp.GPIO.GPIO_ENABLE_W1TC.Set(1 << p)
		}
	}
}

// Get returns the current value of a GPIO pin when the pin is configured as an
// input or as an output.
func (p Pin) Get() bool {
	// See this document for details
	// https://www.espressif.com/sites/default/files/documentation/esp8266-technical_reference_en.pdf

	return esp.GPIO.GPIO_IN.Get()&(1<<p) != 0
}

// Set sets the output value of this pin to high (true) or low (false).
func (p Pin) Set(value bool) {
	if value {
		esp.GPIO.GPIO_OUT_W1TS.Set(1 << p)
	} else {
		esp.GPIO.GPIO_OUT_W1TC.Set(1 << p)
	}
}

// Return the register and mask to enable a given GPIO pin. This can be used to
// implement bit-banged drivers.
//
// Warning: only use this on an output pin!
func (p Pin) PortMaskSet() (*uint32, uint32) {
	return &esp.GPIO.GPIO_OUT_W1TS.Reg, 1 << p
}

// Return the register and mask to disable a given GPIO pin. This can be used to
// implement bit-banged drivers.
//
// Warning: only use this on an output pin!
func (p Pin) PortMaskClear() (*uint32, uint32) {
	return &esp.GPIO.GPIO_OUT_W1TC.Reg, 1 << p
}

var DefaultUART = UART0

// UART0 is a hardware UART that supports both TX and RX.
var UART0 = &_UART0
var _UART0 = UART{Buffer: NewRingBuffer()}

type UART struct {
	Buffer *RingBuffer
}

// Configure the UART baud rate. TX and RX pins are fixed by the hardware so
// cannot be modified and will be ignored.
func (uart *UART) Configure(config UARTConfig) {
	if config.BaudRate == 0 {
		config.BaudRate = 115200
	}
	esp.UART0.UART_CLKDIV.Set(CPUFrequency() / config.BaudRate)
}

// WriteByte writes a single byte to the output buffer. Note that the hardware
// includes a buffer of 128 bytes which will be used first.
func (uart *UART) WriteByte(c byte) error {
	for (esp.UART0.UART_STATUS.Get()>>16)&0xff >= 128 {
		// Wait until the TX buffer has room.
	}
	esp.UART0.UART_FIFO.Set(uint32(c))
	return nil
}
