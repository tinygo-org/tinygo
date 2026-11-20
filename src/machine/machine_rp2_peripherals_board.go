//go:build (rp2040 || rp2350) && !boardpackage

package machine

import (
	"device/rp"
	"runtime/interrupt"
)

var (
	UART0  = &_UART0
	_UART0 = UART{
		Buffer: NewRingBuffer(),
		Bus:    rp.UART0,
	}

	UART1  = &_UART1
	_UART1 = UART{
		Buffer: NewRingBuffer(),
		Bus:    rp.UART1,
	}
)

func init() {
	UART0.Interrupt = interrupt.New(rp.IRQ_UART0_IRQ, _UART0.HandleInterrupt)
	UART1.Interrupt = interrupt.New(rp.IRQ_UART1_IRQ, _UART1.HandleInterrupt)
}

func defaultUARTPins(config *UARTConfig) error {
	if config.TX == 0 && config.RX == 0 {
		config.TX = UART_TX_PIN
		config.RX = UART_RX_PIN
	}
	return nil
}

// I2C on the RP2040/RP2350.
var (
	I2C0  = &_I2C0
	_I2C0 = I2C{
		Bus: rp.I2C0,
	}
	I2C1  = &_I2C1
	_I2C1 = I2C{
		Bus: rp.I2C1,
	}
)

func defaultI2CPins(config *I2CConfig, bus *rp.I2C0_Type) {
	if config.SCL != 0 || config.SDA != 0 {
		return
	}
	switch bus {
	case rp.I2C0:
		config.SCL = I2C0_SCL_PIN
		config.SDA = I2C0_SDA_PIN
	case rp.I2C1:
		config.SCL = I2C1_SCL_PIN
		config.SDA = I2C1_SDA_PIN
	}
}

// SPI on the RP2040/RP2350.
var (
	SPI0 = &SPI{
		Bus: rp.SPI0,
	}
	SPI1 = &SPI{
		Bus: rp.SPI1,
	}
)

func defaultSPIPins(config *SPIConfig, bus *rp.SPI0_Type) {
	zeroPins := config.SCK == 0 && config.SDO == 0
	if !zeroPins || config.SDI != 0 {
		return
	}
	switch bus {
	case rp.SPI0:
		config.SCK = SPI0_SCK_PIN
		config.SDO = SPI0_SDO_PIN
		config.SDI = SPI0_SDI_PIN
	case rp.SPI1:
		config.SCK = SPI1_SCK_PIN
		config.SDO = SPI1_SDO_PIN
		config.SDI = SPI1_SDI_PIN
	}
}

var (
	PWM0 = getPWMGroup(0)
	PWM1 = getPWMGroup(1)
	PWM2 = getPWMGroup(2)
	PWM3 = getPWMGroup(3)
	PWM4 = getPWMGroup(4)
	PWM5 = getPWMGroup(5)
	PWM6 = getPWMGroup(6)
	PWM7 = getPWMGroup(7)
)
