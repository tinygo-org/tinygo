//go:build cortexm && (ra4m1 || py32 || qemu)

package runtime

import "device/arm"

func sleepTicks(d timeUnit) {
	if d <= 0 {
		return
	}

	if hasScheduler {
		if d >= 1000 {
			waitForEvents()
		}
		return
	}

	sleepUntil := ticks() + d
	for ticks() < sleepUntil {
		waitForEvents()
	}
}

func waitForEvents() {
	arm.Asm("wfe")
}
