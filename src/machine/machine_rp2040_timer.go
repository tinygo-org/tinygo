//go:build rp2040
// +build rp2040

package machine

import (
	"device/rp"
	"runtime/volatile"
	"unsafe"
)

const numTimers = 4

type timerType struct {
	timeHW   volatile.Register32
	timeLW   volatile.Register32
	timeHR   volatile.Register32
	timeLR   volatile.Register32
	alarm    [numTimers]volatile.Register32
	armed    volatile.Register32
	timeRawH volatile.Register32
	timeRawL volatile.Register32
	dbgPause volatile.Register32
	pause    volatile.Register32
	intR     volatile.Register32
	intE     volatile.Register32
	intF     volatile.Register32
	intS     volatile.Register32
}

var timer = (*timerType)(unsafe.Pointer(rp.TIMER))

// TimeElapsed returns time elapsed since power up, in microseconds.
func (tmr *timerType) timeElapsed() (us uint64) {
	// Need to make sure that the upper 32 bits of the timer
	// don't change, so read that first
	hi := tmr.timeRawH.Get()
	var lo, nextHi uint32
	for {
		// Read the lower 32 bits
		lo = tmr.timeRawL.Get()
		// Now read the upper 32 bits again and
		// check that it hasn't incremented. If it has, loop around
		// and read the lower 32 bits again to get an accurate value
		nextHi = tmr.timeRawH.Get()
		if hi == nextHi {
			break
		}
		hi = nextHi
	}
	return uint64(hi)<<32 | uint64(lo)
}
