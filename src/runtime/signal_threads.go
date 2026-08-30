//go:build (darwin || (linux && !baremetal && !wasip1 && !wasm_unknown && !wasip2 && !nintendoswitch)) && scheduler.threads

package runtime

import (
	"internal/futex"
)

// Futex the receiver in os/signal waits on. Its value is 1 when the handler
// has stored a signal that the receiver did not read yet.
//
// It cannot wait on signalFutex, because sleepTicks waits on that one too and
// swaps it back to zero, which would lose the wakeup of the receiver.
var signalRecvFutex futex.Futex

// Futex signalWaitUntilIdle waits on. Its value is always zero, so it is only
// a wakeup address. The wait has a timeout because a wake that arrives before
// the wait starts is not remembered.
var signalIdleFutex futex.Futex

// How long signalWaitUntilIdle blocks before rechecking on its own.
const signalIdlePoll = 1e6 // 1ms, in nanoseconds

//go:linkname signal_recv os/signal.signal_recv
func signal_recv() uint32 {
	// Function called from os/signal to get the next received signal.
	for {
		if num, ok := nextReceivedSignal(); ok {
			if receivedSignals.Load() == 0 {
				// That was the last pending signal, so signalWaitUntilIdle can
				// return now.
				signalIdleFutex.WakeAll()
			}
			return num
		}

		// Clear the flag and then read receivedSignals again. The handler
		// stores the signal before the flag, so no wakeup is lost.
		signalRecvFutex.Store(0)
		if receivedSignals.Load() != 0 {
			continue
		}
		signalRecvFutex.Wait(0)
	}
}

//go:linkname signal_waitUntilIdle os/signal.signalWaitUntilIdle
func signal_waitUntilIdle() {
	// Wait until signal_recv has processed all signals. Gosched is a no-op
	// with threads, so this must block.
	for receivedSignals.Load() != 0 {
		signalIdleFutex.WaitUntil(0, signalIdlePoll)
	}
}

// Called from the signal handler to wake signal_recv. An atomic store and a
// futex wake syscall are both safe in a signal handler.
func signalRecvWake() {
	if signalRecvFutex.Swap(1) == 0 {
		// Changed from 0 to 1, so signal_recv may be waiting on it.
		signalRecvFutex.WakeAll()
	}
}

// Reactivate the goroutine waiting for signals, if there are any. There is no
// such goroutine here, because the handler wakes the receiver directly.
func checkSignals() bool {
	return false
}
