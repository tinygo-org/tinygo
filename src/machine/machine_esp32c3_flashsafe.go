//go:build esp32c3

package machine

import "runtime/interrupt"

func flashSafeEnter() (interrupt.State, bool) {
	return interrupt.Disable(), false
}

func flashSafeExit(state interrupt.State, _ bool) {
	interrupt.Restore(state)
}
