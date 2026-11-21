//go:build esp32s3

package machine

import (
	"device/esp"
)

const (
	I2CEXT0_SCL_OUT_IDX = 89
	I2CEXT0_SDA_OUT_IDX = 90
	I2CEXT1_SCL_OUT_IDX = 91
	I2CEXT1_SDA_OUT_IDX = 92
)

var (
	I2C0 = &I2C{Bus: esp.I2C0}
	I2C1 = &I2C{Bus: esp.I2C1}
)

func i2cBusFor(bus *esp.I2C_Type) i2cBus {
	switch bus {
	case esp.I2C0:
		return i2cBus{sclOut: I2CEXT0_SCL_OUT_IDX, sdaOut: I2CEXT0_SDA_OUT_IDX}
	case esp.I2C1:
		return i2cBus{sclOut: I2CEXT1_SCL_OUT_IDX, sdaOut: I2CEXT1_SDA_OUT_IDX}
	default:
		return i2cBus{}
	}
}

// enableI2CClock enables the clock of the I2C bus and resets it.
func enableI2CClock(bus *esp.I2C_Type) {
	switch bus {
	case esp.I2C0:
		esp.SYSTEM.SetPERIP_RST_EN0_I2C_EXT0_RST(1)
		esp.SYSTEM.SetPERIP_CLK_EN0_I2C_EXT0_CLK_EN(1)
		esp.SYSTEM.SetPERIP_RST_EN0_I2C_EXT0_RST(0)
	case esp.I2C1:
		esp.SYSTEM.SetPERIP_RST_EN0_I2C_EXT1_RST(1)
		esp.SYSTEM.SetPERIP_CLK_EN0_I2C_EXT1_CLK_EN(1)
		esp.SYSTEM.SetPERIP_RST_EN0_I2C_EXT1_RST(0)
	}
}
