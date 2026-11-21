//go:build esp32p4

package machine

import (
	"device/esp"
)

// GPIO matrix signals, see ESP-IDF components/soc/esp32p4/include/soc/gpio_sig_map.h.
const (
	I2C0_SCL_PAD_OUT_IDX = 68
	I2C0_SDA_PAD_OUT_IDX = 69
	I2C1_SCL_PAD_OUT_IDX = 70
	I2C1_SDA_PAD_OUT_IDX = 71
)

var (
	I2C0 = &I2C{Bus: esp.I2C0}
	I2C1 = &I2C{Bus: esp.I2C1}
)

func i2cBusFor(bus *esp.I2C_Type) i2cBus {
	switch bus {
	case esp.I2C0:
		return i2cBus{sclOut: I2C0_SCL_PAD_OUT_IDX, sdaOut: I2C0_SDA_PAD_OUT_IDX}
	case esp.I2C1:
		return i2cBus{sclOut: I2C1_SCL_PAD_OUT_IDX, sdaOut: I2C1_SDA_PAD_OUT_IDX}
	default:
		return i2cBus{}
	}
}

// enableI2CClock enables the clock of the I2C bus and resets it.
func enableI2CClock(bus *esp.I2C_Type) {
	switch bus {
	case esp.I2C0:
		esp.HP_SYS_CLKRST.SetSOC_CLK_CTRL2_REG_I2C0_APB_CLK_EN(1)
		esp.HP_SYS_CLKRST.SetHP_RST_EN1_REG_RST_EN_I2C0(1)
		esp.HP_SYS_CLKRST.SetHP_RST_EN1_REG_RST_EN_I2C0(0)
		esp.HP_SYS_CLKRST.SetPERI_CLK_CTRL10_REG_I2C0_CLK_SRC_SEL(i2cClockSource)
		esp.HP_SYS_CLKRST.SetPERI_CLK_CTRL10_REG_I2C0_CLK_EN(1)
	case esp.I2C1:
		esp.HP_SYS_CLKRST.SetSOC_CLK_CTRL2_REG_I2C1_APB_CLK_EN(1)
		esp.HP_SYS_CLKRST.SetHP_RST_EN1_REG_RST_EN_I2C1(1)
		esp.HP_SYS_CLKRST.SetHP_RST_EN1_REG_RST_EN_I2C1(0)
		esp.HP_SYS_CLKRST.SetPERI_CLK_CTRL10_REG_I2C1_CLK_SRC_SEL(i2cClockSource)
		esp.HP_SYS_CLKRST.SetPERI_CLK_CTRL10_REG_I2C1_CLK_EN(1)
	}
}

// setI2CClockDiv sets the I2C source clock divider, which is in HP_SYS_CLKRST on the P4.
// See ESP-IDF hal/esp32p4/include/hal/i2c_ll.h i2c_ll_master_set_bus_timing.
func setI2CClockDiv(bus *esp.I2C_Type, div uint32) {
	switch bus {
	case esp.I2C0:
		esp.HP_SYS_CLKRST.SetPERI_CLK_CTRL10_REG_I2C0_CLK_DIV_NUM(div)
		esp.HP_SYS_CLKRST.SetPERI_CLK_CTRL10_REG_I2C0_CLK_DIV_NUMERATOR(0)
		esp.HP_SYS_CLKRST.SetPERI_CLK_CTRL10_REG_I2C0_CLK_DIV_DENOMINATOR(0)
	case esp.I2C1:
		esp.HP_SYS_CLKRST.SetPERI_CLK_CTRL11_REG_I2C1_CLK_DIV_NUM(div)
		esp.HP_SYS_CLKRST.SetPERI_CLK_CTRL11_REG_I2C1_CLK_DIV_NUMERATOR(0)
		esp.HP_SYS_CLKRST.SetPERI_CLK_CTRL11_REG_I2C1_CLK_DIV_DENOMINATOR(0)
	}
}
