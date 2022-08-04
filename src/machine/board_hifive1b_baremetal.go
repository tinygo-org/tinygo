//go:build fe310 && hifive1b
// +build fe310,hifive1b

package machine

import "tinygo.org/x/device/sifive"

// SPI on the HiFive1.
var (
	SPI1 = SPI{
		Bus: sifive.QSPI1,
	}
)
