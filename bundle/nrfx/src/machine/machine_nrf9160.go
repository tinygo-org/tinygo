//go:build nrf9160
// +build nrf9160

package machine

import (
	"device/nrf"
	"runtime/interrupt"
)

// Hardware pins
const (
	P0_00 Pin = 0
	P0_01 Pin = 1
	P0_02 Pin = 2
	P0_03 Pin = 3
	P0_04 Pin = 4
	P0_05 Pin = 5
	P0_06 Pin = 6
	P0_07 Pin = 7
	P0_08 Pin = 8
	P0_09 Pin = 9
	P0_10 Pin = 10
	P0_11 Pin = 11
	P0_12 Pin = 12
	P0_13 Pin = 13
	P0_14 Pin = 14
	P0_15 Pin = 15
	P0_16 Pin = 16
	P0_17 Pin = 17
	P0_18 Pin = 18
	P0_19 Pin = 19
	P0_20 Pin = 20
	P0_21 Pin = 21
	P0_22 Pin = 22
	P0_23 Pin = 23
	P0_24 Pin = 24
	P0_25 Pin = 25
	P0_26 Pin = 26
	P0_27 Pin = 27
	P0_28 Pin = 28
	P0_29 Pin = 29
	P0_30 Pin = 30
	P0_31 Pin = 31
)

// Get peripheral and pin number for this GPIO pin.
func (p Pin) getPortPin() (*nrf.GPIO_Type, uint32) {
	return nrf.P0_S, uint32(p)
}

// UART
var (
	// UART0 is the hardware EasyDMA UART on the NRF SoC.
	UART0  = &_UART0
	UART1  = &_UART1
	_UART0 = UARTE{UARTE_Type: nrf.UARTE0_S, Buffer: NewRingBuffer()}
	_UART1 = UARTE{UARTE_Type: nrf.UARTE1_S, Buffer: NewRingBuffer()}

	B_115200 = nrf.UARTE_BAUDRATE_BAUDRATE_Baud115200
)

func init() {
	UART0.Interrupt = interrupt.New(nrf.IRQ_UARTE0_S, _UART0.handleInterrupt)
	UART1.Interrupt = interrupt.New(nrf.IRQ_UARTE1_S, _UART1.handleInterrupt)
}

// I2C on the NRF.
type I2C struct {
	Bus *nrf.TWIM_Type
}

// There are 2 I2C interfaces on the NRF.
var (
	I2C0 = &_I2C0
	I2C1 = &_I2C1
	_I2C0 = I2C{Bus: nrf.TWIM0_S}
	_I2C1 = I2C{Bus: nrf.TWIM1_S}
)

func (i2c *I2C) setPins(scl, sda Pin) {
	i2c.Bus.PSEL.SCL.Set(uint32(scl))
	i2c.Bus.PSEL.SDA.Set(uint32(sda))
}

// PWM
var (
	PWM0 = &PWM{PWM: nrf.PWM0_S}
	PWM1 = &PWM{PWM: nrf.PWM1_S}
	PWM2 = &PWM{PWM: nrf.PWM2_S}
	PWM3 = &PWM{PWM: nrf.PWM3_S}
)

func (i2c *I2C) Tx(addr uint16, w, r []byte) (err error) {
	return nil
}

/*
// writeByte writes a single byte to the I2C bus and waits for confirmation.
func (i2c *I2C) writeByte(data byte) error {
	i2c.Bus.TXD.Set(uint32(data))
	for i2c.Bus.EVENTS_TXDSENT.Get() == 0 {
		if e := i2c.Bus.EVENTS_ERROR.Get(); e != 0 {
			i2c.Bus.EVENTS_ERROR.Set(0)
			return errI2CBusError
		}
	}
	i2c.Bus.EVENTS_TXDSENT.Set(0)
	return nil
}

// readByte reads a single byte from the I2C bus when it is ready.
func (i2c *I2C) readByte() (byte, error) {
	for i2c.Bus.EVENTS_RXDREADY.Get() == 0 {
		if e := i2c.Bus.EVENTS_ERROR.Get(); e != 0 {
			i2c.Bus.EVENTS_ERROR.Set(0)
			return 0, errI2CBusError
		}
	}
	i2c.Bus.EVENTS_RXDREADY.Set(0)
	return byte(i2c.Bus.RXD.Get()), nil
}
*/

// There are 3 SPI interfaces on the NRF528xx.
var (
	SPI0 = SPI{Bus: nrf.SPIM0_S, buf: new([1]byte)}
	SPI1 = SPI{Bus: nrf.SPIM1_S, buf: new([1]byte)}
	SPI2 = SPI{Bus: nrf.SPIM2_S, buf: new([1]byte)}
)
