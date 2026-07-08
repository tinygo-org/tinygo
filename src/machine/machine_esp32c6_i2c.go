//go:build esp32c6

package machine

import (
	"device/esp"
)

// GPIO matrix signal indices for I2C0 on ESP32-C6.
const (
	I2CEXT0_SCL_OUT_IDX = 45
	I2CEXT0_SDA_OUT_IDX = 46
)

var (
	I2C0 = &I2C{
		Bus:     esp.I2C0,
		funcSCL: I2CEXT0_SCL_OUT_IDX,
		funcSDA: I2CEXT0_SDA_OUT_IDX,
		useExt1: false,
	}
)

// enableI2C0PeriphClock enables the I2C0 peripheral clock via PCR.
func enableI2C0PeriphClock() {
	// Enable the APB/bus clock for the I2C0 registers and pulse the reset.
	esp.PCR.SetI2C0_CONF_I2C0_CLK_EN(1)
	esp.PCR.SetI2C0_CONF_I2C0_RST_EN(1)
	esp.PCR.SetI2C0_CONF_I2C0_RST_EN(0)

	// On the ESP32-C6 the I2C functional (source) clock is gated and selected
	// in PCR.
	// Select the XTAL (40 MHz) source and enable the clock gate, otherwise SCL
	// is never driven and no bytes are clocked onto the bus.
	esp.PCR.SetI2C_SCLK_CONF_I2C_SCLK_SEL(i2cClkSource)
	esp.PCR.SetI2C_SCLK_CONF_I2C_SCLK_EN(1)
}
