//go:build (rp2040 || rp2350) && boardpackage

package machine

import "device/rp"

func defaultI2CPins(config *I2CConfig, bus *rp.I2C0_Type) {}

func defaultSPIPins(config *SPIConfig, bus *rp.SPI0_Type) {}
