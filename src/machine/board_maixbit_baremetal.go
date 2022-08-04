//go:build k210 && maixbit
// +build k210,maixbit

package machine

import "tinygo.org/x/device/kendryte"

// SPI on the MAix Bit.
var (
	SPI0 = SPI{
		Bus: kendryte.SPI0,
	}
	SPI1 = SPI{
		Bus: kendryte.SPI1,
	}
)
