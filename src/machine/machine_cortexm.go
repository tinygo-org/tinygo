//go:build cortexm

package machine

import "device/arm"

// CPUReset performs a hard system reset.
func CPUReset() {
	arm.SystemReset()
}
