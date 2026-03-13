//go:build esp32c3 || esp32s3

package machine

import "device/esp"

// enableI2C0PeriphClock enables the I2C0 peripheral clock via SYSTEM.
func enableI2C0PeriphClock() {
	esp.SYSTEM.SetPERIP_RST_EN0_I2C_EXT0_RST(1)
	esp.SYSTEM.SetPERIP_CLK_EN0_I2C_EXT0_CLK_EN(1)
	esp.SYSTEM.SetPERIP_RST_EN0_I2C_EXT0_RST(0)
}

// enableLEDCPeriphClock enables the LEDC peripheral clock via SYSTEM.
func enableLEDCPeriphClock() {
	esp.SYSTEM.SetPERIP_RST_EN0_LEDC_RST(1)
	esp.SYSTEM.SetPERIP_CLK_EN0_LEDC_CLK_EN(1)
	esp.SYSTEM.SetPERIP_RST_EN0_LEDC_RST(0)
}
