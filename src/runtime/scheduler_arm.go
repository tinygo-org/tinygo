// +build arm,!avr,baremetal
// +build atsamd21

package runtime

import (
	"device/arm"
	"internal/task"
)

var intq task.InterruptQueue

func pushInterrupt(t *task.Task) {
	intq.Push(t)
}

func poll() bool {
	var found bool

	// Check the interrupt queue.
	if !intq.Empty() {
		intq.AppendTo(&runqueue)
		found = true
	}

	// Check for device-specific events.
	found = devicePoll() || found

	return found
}

func wait() {
	arm.Asm("wfi")
}
