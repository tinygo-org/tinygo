//go:build esp32s3

package machine

import "device/esp"

// ReadTemperature reads the on-chip temperature sensor (TSENS) and returns
// a value in millicelsius (°C × 1000). Uses the default measurement range
// (offset = 0, approximately −10 °C to 80 °C, ±3 °C accuracy).
//
// The conversion uses the same formula as ESP-IDF with no eFuse calibration:
//
//	T = (0.4386 × raw − 20.52) °C
func ReadTemperature() int32 {
	// Enable TSENS peripheral clock.
	esp.SENS.SetSAR_PERI_CLK_GATE_CONF_TSENS_CLK_EN(1)

	// Set clock divider to default (6).
	esp.SENS.SetSAR_TSENS_CTRL_SAR_TSENS_CLK_DIV(6)

	// Power up the temperature sensor.
	esp.SENS.SetSAR_TSENS_CTRL_SAR_TSENS_POWER_UP_FORCE(1)
	esp.SENS.SetSAR_TSENS_CTRL2_SAR_TSENS_XPD_FORCE(1)
	esp.SENS.SetSAR_TSENS_CTRL_SAR_TSENS_POWER_UP(1)

	// Trigger a conversion.
	esp.SENS.SetSAR_TSENS_CTRL_SAR_TSENS_DUMP_OUT(1)

	// Wait for data ready.
	for esp.SENS.GetSAR_TSENS_CTRL_SAR_TSENS_READY() == 0 {
	}

	// Read the 8-bit raw value.
	raw := int32(esp.SENS.GetSAR_TSENS_CTRL_SAR_TSENS_OUT())

	// Stop the conversion.
	esp.SENS.SetSAR_TSENS_CTRL_SAR_TSENS_DUMP_OUT(0)

	// Convert to millicelsius using the ESP-IDF integer formula (offset=0):
	//   T_celsius     = (4386 * raw - 205200) / 10000
	//   T_millicelsius = (4386 * raw - 205200) / 10
	return (4386*raw - 205200) / 10
}
