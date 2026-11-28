//go:build cortexm && !rp2040 && !rp2350

package machine

import "device/arm"

// CPUReset performs a hard system reset.
func CPUReset() {
	arm.SystemReset()
}
