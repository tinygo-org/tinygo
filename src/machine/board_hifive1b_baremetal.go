// +build fe310,hifive1b

package machine

import "github.com/sago35/device/sifive"

// SPI on the HiFive1.
var (
	SPI1 = SPI{
		Bus: sifive.QSPI1,
	}
)
