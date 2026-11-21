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
		// Enable the APB/bus clock for the I2C0 registers and pulse the reset.
		esp.PCR.SetI2C0_CONF_I2C0_CLK_EN(1)
		esp.PCR.SetI2C0_CONF_I2C0_RST_EN(1)
		esp.PCR.SetI2C0_CONF_I2C0_RST_EN(0)

		// On the ESP32-C6 the I2C functional (source) clock is gated and selected
		// in PCR.
		// Select the XTAL (40 MHz) source and enable the clock gate, otherwise SCL
		// is never driven and no bytes are clocked onto the bus.
		esp.PCR.SetI2C_SCLK_CONF_I2C_SCLK_SEL(i2cClockSource)
		esp.PCR.SetI2C_SCLK_CONF_I2C_SCLK_EN(1)
	}
}

// setI2CClockDiv sets the I2C source clock divider, which is in PCR on the C6.
// See ESP-IDF hal/esp32c6/include/hal/i2c_ll.h i2c_ll_master_set_bus_timing.
func setI2CClockDiv(bus *esp.I2C_Type, div uint32) {
	esp.PCR.SetI2C_SCLK_CONF_I2C_SCLK_DIV_NUM(div)
	esp.PCR.SetI2C_SCLK_CONF_I2C_SCLK_DIV_A(0)
	esp.PCR.SetI2C_SCLK_CONF_I2C_SCLK_DIV_B(0)
}
