//go:build esp32c3 && !m5stamp_c3

package machine

import (
	"device/esp"
)

const (
	I2CEXT0_SCL_OUT_IDX = 53
	I2CEXT0_SDA_OUT_IDX = 54
)

var (
	I2C0 = &I2C{
		Bus:     esp.I2C0,
		funcSCL: I2CEXT0_SCL_OUT_IDX,
		funcSDA: I2CEXT0_SDA_OUT_IDX,
		useExt1: false,
	}
)
