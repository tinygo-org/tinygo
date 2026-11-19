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
	I2C0 = &I2C{
		Bus:     esp.I2C0,
		funcSCL: I2C0_SCL_PAD_OUT_IDX,
		funcSDA: I2C0_SDA_PAD_OUT_IDX,
		useExt1: false,
	}
	I2C1 = &I2C{
		Bus:     esp.I2C1,
		funcSCL: I2C1_SCL_PAD_OUT_IDX,
		funcSDA: I2C1_SDA_PAD_OUT_IDX,
		useExt1: true,
	}
)

// enableI2C0PeriphClock enables the I2C0 peripheral APB clock and resets the
// peripheral, then enables the functional (source) clock via HP_SYS_CLKRST.
func enableI2C0PeriphClock() {
	esp.HP_SYS_CLKRST.SetSOC_CLK_CTRL2_REG_I2C0_APB_CLK_EN(1)
	esp.HP_SYS_CLKRST.SetHP_RST_EN1_REG_RST_EN_I2C0(1)
	esp.HP_SYS_CLKRST.SetHP_RST_EN1_REG_RST_EN_I2C0(0)
	esp.HP_SYS_CLKRST.SetPERI_CLK_CTRL10_REG_I2C0_CLK_SRC_SEL(0) // 0 = XTAL
	esp.HP_SYS_CLKRST.SetPERI_CLK_CTRL10_REG_I2C0_CLK_EN(1)
}

// initI2CExt1Clock enables the I2C1 peripheral APB clock and resets the
// peripheral, then enables the functional (source) clock via HP_SYS_CLKRST.
func initI2CExt1Clock() {
	esp.HP_SYS_CLKRST.SetSOC_CLK_CTRL2_REG_I2C1_APB_CLK_EN(1)
	esp.HP_SYS_CLKRST.SetHP_RST_EN1_REG_RST_EN_I2C1(1)
	esp.HP_SYS_CLKRST.SetHP_RST_EN1_REG_RST_EN_I2C1(0)
	esp.HP_SYS_CLKRST.SetPERI_CLK_CTRL10_REG_I2C1_CLK_SRC_SEL(0) // 0 = XTAL
	esp.HP_SYS_CLKRST.SetPERI_CLK_CTRL10_REG_I2C1_CLK_EN(1)
}
