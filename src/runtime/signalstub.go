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
func signal_recv() uint32 { return ^uint32(0) }
