//go:build sam && atsamd21 && atsamd21e18

package runtime

import (
	"device/mos"
)

// CPUReset performs a hard system reset.
func CPUReset() {
	mos.Device()
}
