//go:build tinygo.wasm || baremetal

package runtime

// Some platforms don't support Unix signals (and never will), so we need to
// stub the signal functions.

//go:linkname signal_disable os/signal.signal_disable
func signal_disable(uint32) {}

//go:linkname signal_enable os/signal.signal_enable
func signal_enable(uint32) {}

//go:linkname signal_ignore os/signal.signal_ignore
func signal_ignore(uint32) {}

//go:linkname signal_waitUntilIdle os/signal.signalWaitUntilIdle
func signal_waitUntilIdle() {}

//go:linkname signal_recv os/signal.signal_recv
func signal_recv() uint32 {
	// Signals can never arrive on these platforms, so block forever, like the
	// real implementation does while no signal is pending. os/signal.loop
	// calls this function in a tight loop; returning immediately would make
	// that goroutine spin, which starves cooperative schedulers: on wasm the
	// program never yields back to the host again after signal.Notify.
	deadlock()
	return 0
}
