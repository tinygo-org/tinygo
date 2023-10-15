//go:build !baremetal || atmega || esp32 || fe310 || k210 || nrf || (nxp && !mk66f18) || rp2040 || sam || (stm32 && !stm32f7x2 && !stm32l5x2)

package machine

import "errors"

// SPI phase and polarity configs CPOL and CPHA
const (
	Mode0 = 0
	Mode1 = 1
	Mode2 = 2
	Mode3 = 3
)

var (
	ErrTxInvalidSliceSize      = errors.New("SPI write and read slices must be same size")
	errSPIInvalidMachineConfig = errors.New("SPI port was not configured properly by the machine")
)

// If you are getting a compile error on this line please check to see you've
// correctly implemented the methods on the SPI type. They must match
// the interface method signatures type to type perfectly.
// If not implementing the SPI type please remove your target from the build tags
// at the top of this file.
var _ interface { // 2
	Configure(config SPIConfig) error
	Tx(w, r []byte) error
	Transfer(w byte) (byte, error)
} = (*SPI)(nil)
