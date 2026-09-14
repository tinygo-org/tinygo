package main

// Verify that signal.Notify does not wedge the scheduler on platforms where
// signals can never arrive (wasm and baremetal, see runtime/signalstub.go).
// The signal watcher goroutine that Notify starts must block forever instead
// of spinning: on a cooperative scheduler a spinning goroutine starves every
// other goroutine, so the sleep below would never return.

import (
	"os"
	"os/signal"
	"time"
)

func main() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	// Yield so the signal watcher goroutine runs (and, before the fix, takes
	// over the scheduler forever).
	time.Sleep(10 * time.Millisecond)
	println("done")
}
