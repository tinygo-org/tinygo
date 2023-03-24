//go:build cortexm

package machine

import "device/arm"

// SystemReset performs a hard system reset.
func SystemReset() {
	arm.SystemReset()
}
