package msc

import (
	"runtime/interrupt"
	"sync/atomic"
)

// spinLock guards state shared by the USB interrupt and processTasks, also when
// processTasks runs on the other core with scheduler=cores.
type spinLock struct {
	held atomic.Uint32
}

func (l *spinLock) lock() interrupt.State {
	for {
		state := interrupt.Disable()
		if l.held.CompareAndSwap(0, 1) {
			return state
		}
		interrupt.Restore(state)
	}
}

func (l *spinLock) unlock(state interrupt.State) {
	l.held.Store(0)
	interrupt.Restore(state)
}
