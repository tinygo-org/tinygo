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
	I2C0 = &I2C{
		Bus:     esp.I2C0,
		funcSCL: I2CEXT0_SCL_OUT_IDX,
		funcSDA: I2CEXT0_SDA_OUT_IDX,
		useExt1: false,
	}
	I2C1 = &I2C{
		Bus:     esp.I2C1,
		funcSCL: I2CEXT1_SCL_OUT_IDX,
		funcSDA: I2CEXT1_SDA_OUT_IDX,
		useExt1: true,
	}
)

func initI2CExt1Clock() {
	esp.SYSTEM.SetPERIP_RST_EN0_I2C_EXT1_RST(1)
	esp.SYSTEM.SetPERIP_CLK_EN0_I2C_EXT1_CLK_EN(1)
	esp.SYSTEM.SetPERIP_RST_EN0_I2C_EXT1_RST(0)
}
