//go:build cortexm

package machine

import "device/mos"

// CPUReset performs a hard system reset.
func CPUReset() {
	mos.Device()
}
