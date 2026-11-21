//go:build esp32c3

package machine

import (
	"device/esp"
)

const (
	I2CEXT0_SCL_OUT_IDX = 53
	I2CEXT0_SDA_OUT_IDX = 54
)

var (
	I2C0 = &I2C{Bus: esp.I2C0}
)

func i2cBusFor(bus *esp.I2C_Type) i2cBus {
	switch bus {
	case esp.I2C0:
		return i2cBus{sclOut: I2CEXT0_SCL_OUT_IDX, sdaOut: I2CEXT0_SDA_OUT_IDX}
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
	}
}

// setI2CClockDiv sets the I2C source clock divider. See ESP-IDF
// hal/esp32c3/include/hal/i2c_ll.h i2c_ll_master_set_bus_timing.
func setI2CClockDiv(bus *esp.I2C_Type, div uint32) {
	bus.SetCLK_CONF_SCLK_DIV_NUM(div)
}
