//go:build tkey

package machine

import (
	"device/tkey"
	"errors"
	"strconv"
)

const deviceName = "TKey"

// GPIO pins modes are only here to match the Pin interface.
// The actual configuration is fixed in the hardware.
const (
	PinOutput PinMode = iota
	PinInput
	PinInputPullup
	PinInputPulldown
)

const (
	LED_BLUE  = Pin(tkey.TK1_MMIO_TK1_LED_B_BIT)
	LED_GREEN = Pin(tkey.TK1_MMIO_TK1_LED_G_BIT)
	LED_RED   = Pin(tkey.TK1_MMIO_TK1_LED_R_BIT)

	LED = LED_GREEN

	TKEY_TOUCH = Pin(3) // 3 is unused, but we need a value here to match the Pin interface.
	BUTTON     = TKEY_TOUCH

	GPIO1 = Pin(tkey.TK1_MMIO_TK1_GPIO1_BIT + 8)
	GPIO2 = Pin(tkey.TK1_MMIO_TK1_GPIO2_BIT + 8)
	GPIO3 = Pin(tkey.TK1_MMIO_TK1_GPIO3_BIT + 8)
	GPIO4 = Pin(tkey.TK1_MMIO_TK1_GPIO4_BIT + 8)
)

var touchConfig, gpio1Config, gpio2Config PinConfig

// No config needed for TKey, just to match the Pin interface.
func (p Pin) Configure(config PinConfig) {
	switch p {
	case BUTTON:
		touchConfig = config

		// Clear any pending touch events.
		tkey.TOUCH.STATUS.Set(0)
	case GPIO1:
		gpio1Config = config
	case GPIO2:
		gpio2Config = config
	}
}

// Set pin to high or low.
func (p Pin) Set(high bool) {
	switch p {
	case LED_BLUE, LED_GREEN, LED_RED:
		if high {
			tkey.TK1.LED.SetBits(1 << uint(p))
		} else {
			tkey.TK1.LED.ClearBits(1 << uint(p))
		}
	case GPIO3, GPIO4:
		if high {
			tkey.TK1.GPIO.SetBits(1 << uint(p-8))
		} else {
			tkey.TK1.GPIO.ClearBits(1 << uint(p-8))
		}
	}
}

// Get returns the current value of a pin.
func (p Pin) Get() bool {
	pushed := false
	mode := PinInput

	switch p {
	case BUTTON:
		mode = touchConfig.Mode
		if tkey.TOUCH.STATUS.HasBits(1) {
			tkey.TOUCH.STATUS.Set(0)
			pushed = true
		}
	case GPIO1:
		mode = gpio1Config.Mode
		pushed = tkey.TK1.GPIO.HasBits(1 << uint(p-8))
	case GPIO2:
		mode = gpio2Config.Mode
		pushed = tkey.TK1.GPIO.HasBits(1 << uint(p-8))
	case GPIO3, GPIO4:
		mode = PinOutput
		pushed = tkey.TK1.GPIO.HasBits(1 << uint(p-8))
	case LED_BLUE, LED_GREEN, LED_RED:
		mode = PinOutput
		pushed = tkey.TK1.LED.HasBits(1 << uint(p))
	}

	switch mode {
	case PinInputPullup:
		return !pushed
	case PinInput, PinInputPulldown, PinOutput:
		return pushed
	}

	return false
}

type UART struct {
	Bus *tkey.UART_Type
}

var (
	DefaultUART = UART0
	UART0       = &_UART0
	_UART0      = UART{Bus: tkey.UART}
)

// The TKey UART is fixed at 62500 baud, 8N1.
func (uart *UART) Configure(config UARTConfig) error {
	if !(config.BaudRate == 62500 || config.BaudRate == 0) {
		return errors.New("uart: only 62500 baud rate is supported")
	}

	return nil
}

// Write a slice of data bytes to the UART.
func (uart *UART) Write(data []byte) (n int, err error) {
	for _, c := range data {
		if err := uart.WriteByte(c); err != nil {
			return n, err
		}
	}
	return len(data), nil
}

// WriteByte writes a byte of data to the UART.
func (uart *UART) WriteByte(c byte) error {
	for uart.Bus.TX_STATUS.Get() == 0 {
	}

	uart.Bus.TX_DATA.Set(uint32(c))

	return nil
}

// Buffered returns the number of bytes buffered in the UART.
func (uart *UART) Buffered() int {
	return int(uart.Bus.RX_BYTES.Get())
}

// ReadByte reads a byte of data from the UART.
func (uart *UART) ReadByte() (byte, error) {
	for uart.Bus.RX_STATUS.Get() == 0 {
	}

	return byte(uart.Bus.RX_DATA.Get()), nil
}

// DTR is not available on the TKey.
func (uart *UART) DTR() bool {
	return false
}

// RTS is not available on the TKey.
func (uart *UART) RTS() bool {
	return false
}

// GetRNG returns 32 bits of cryptographically secure random data
func GetRNG() (uint32, error) {
	for tkey.TRNG.STATUS.Get() == 0 {
	}

	return uint32(tkey.TRNG.ENTROPY.Get()), nil
}

// DesignName returns the FPGA design name.
func DesignName() (string, string) {
	n0 := tkey.TK1.NAME0.Get()
	name0 := string([]byte{byte(n0 >> 24), byte(n0 >> 16), byte(n0 >> 8), byte(n0)})
	n1 := tkey.TK1.NAME1.Get()
	name1 := string([]byte{byte(n1 >> 24), byte(n1 >> 16), byte(n1 >> 8), byte(n1)})

	return name0, name1
}

// DesignVersion returns the FPGA design version.
func DesignVersion() string {
	version := tkey.TK1.VERSION.Get()

	return strconv.Itoa(int(version))
}

// CDI returns 8 words of Compound Device Identifier (CDI) generated and written by the firmware when the application is loaded.
func CDI() []byte {
	cdi := make([]byte, 32)
	for i := 0; i < 8; i++ {
		c := tkey.TK1.CDI_FIRST[i].Get()
		cdi[i*4] = byte(c >> 24)
		cdi[i*4+1] = byte(c >> 16)
		cdi[i*4+2] = byte(c >> 8)
		cdi[i*4+3] = byte(c)
	}
	return cdi
}

// UDI returns 2 words of Unique Device Identifier (UDI). Only available in firmware mode.
func UDI() []byte {
	udi := make([]byte, 8)
	for i := 0; i < 2; i++ {
		c := tkey.TK1.UDI_FIRST[i].Get()
		udi[i*4] = byte(c >> 24)
		udi[i*4+1] = byte(c >> 16)
		udi[i*4+2] = byte(c >> 8)
		udi[i*4+3] = byte(c)
	}
	return udi
}

// UDS returns 8 words of Unique Device Secret. Part of the FPGA design, changed when provisioning a TKey.
// Only available in firmware mode. UDS is only readable once per power cycle.
func UDS() []byte {
	uds := make([]byte, 32)
	for i := 0; i < 8; i++ {
		c := tkey.UDS.DATA[i].Get()
		uds[i*4] = byte(c >> 24)
		uds[i*4+1] = byte(c >> 16)
		uds[i*4+2] = byte(c >> 8)
		uds[i*4+3] = byte(c)
	}
	return uds
}
