//go:build rp2350 && !rp2350b

package machine

import "device/rp"

// Analog pin base and internal temperature channel on RP2350A.
const (
	adc0Pin Pin = GPIO26

	// fifth ADC channel.
	thermADC = 30
)

// validPins confirms that the SPI pin selection is a legitimate one
// for the the 2350a chip.
func (spi *SPI) validPins(config SPIConfig) error {
	var okSDI, okSDO, okSCK bool
	switch spi.Bus {
	case rp.SPI0:
		okSDI = config.SDI == NoPin || config.SDI == 0 || config.SDI == 4 || config.SDI == 16 || config.SDI == 20
		okSDO = config.SDO == 3 || config.SDO == 7 || config.SDO == 19 || config.SDO == 23
		okSCK = config.SCK == 2 || config.SCK == 6 || config.SCK == 18 || config.SCK == 22
	case rp.SPI1:
		okSDI = config.SDI == NoPin || config.SDI == 8 || config.SDI == 12 || config.SDI == 24 || config.SDI == 28
		okSDO = config.SDO == 11 || config.SDO == 15 || config.SDO == 27
		okSCK = config.SCK == 10 || config.SCK == 14 || config.SCK == 26
	}
	switch {
	case !okSDI:
		return errSPIInvalidSDI
	case !okSDO:
		return errSPIInvalidSDO
	case !okSCK:
		return errSPIInvalidSCK
	}
	return nil
}
