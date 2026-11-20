//go:build (rp2040 || rp2350) && machine.uart && boardpackage

package machine

import "errors"

var errUARTPinsRequired = errors.New("uart: pins required")

func defaultUARTPins(config *UARTConfig) error {
	bothZero := config.TX == 0 && config.RX == 0
	bothNoPin := config.TX == NoPin && config.RX == NoPin
	if bothZero || bothNoPin {
		return errUARTPinsRequired
	}
	return nil
}
