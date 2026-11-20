//go:build (rp2040 || rp2350) && boardpackage

package machine

import "device/rp"

func defaultUARTPins(config *UARTConfig) error {
	bothZero := config.TX == 0 && config.RX == 0
	bothNoPin := config.TX == NoPin && config.RX == NoPin
	if bothZero || bothNoPin {
		return ErrInvalidOutputPin
	}
	return nil
}

func defaultI2CPins(config *I2CConfig, bus *rp.I2C0_Type) {}

func defaultSPIPins(config *SPIConfig, bus *rp.SPI0_Type) {}
