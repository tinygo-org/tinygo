//go:build (darwin || (linux && !baremetal && !wasip1 && !wasm_unknown && !wasip2 && !nintendoswitch)) && !scheduler.threads

package runtime

import (
	"internal/task"
	"sync/atomic"
)

// The goroutine inside os/signal that reads signals is an ordinary task here.
// It is parked with task.Pause and resumed from checkSignals below.

// Task waiting for a signal to arrive, or nil if it is running or there are no
// signals.
var signalRecvWaiter atomic.Pointer[task.Task]

//go:linkname signal_recv os/signal.signal_recv
func signal_recv() uint32 {
	// Function called from os/signal to get the next received signal.
	for {
		if num, ok := nextReceivedSignal(); ok {
			return num
		}

		// There are no signals to receive. Sleep until there are.
		if signalRecvWaiter.Swap(task.Current()) != nil {
			// We expect only a single goroutine to call signal_recv.
			runtimeFatal("signal_recv called concurrently")
		}
		task.Pause()
	}
}

//go:linkname signal_waitUntilIdle os/signal.signalWaitUntilIdle
func signal_waitUntilIdle() {
	// Wait until signal_recv has processed all signals. Yielding is enough:
	// the scheduler runs signal_recv, which is the only thing that empties
	// receivedSignals.
	for receivedSignals.Load() != 0 {
		Gosched()
	}
}

// Called from the signal handler. The waiting task is resumed by checkSignals
// instead, from the scheduler, so there is nothing to do here.
func signalRecvWake() {
}

// Reactivate the goroutine waiting for signals, if there are any.
// Return true if it was reactivated (and therefore the scheduler should run
// again), and false otherwise.
func checkSignals() bool {
	if receivedSignals.Load() != 0 {
		if waiter := signalRecvWaiter.Swap(nil); waiter != nil {
			scheduleTask(waiter)
			return true
		}
	}
	return false
}
