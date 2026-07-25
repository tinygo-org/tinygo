// This example demonstrates goroutine core pinning on multi-core systems (RP2040/RP2350).
// It shows how to pin goroutines to specific CPU cores and verify their execution.

//go:build rp2040 || rp2350

package main

import (
	"machine"
	"runtime"
	"time"
)

func main() {
	time.Sleep(5 * time.Second)
	println("=== Core Pinning Example ===")
	println("Number of CPU cores:", runtime.NumCPU())
	println("[main] Main starting on core:", machine.CurrentCore())
	println()

	// Example 1: Pin using standard Go API (LockOSThread)
	// This pins to whichever core this goroutine is currently running on
	runtime.LockOSThread()
	println("[main] Pinned using runtime.LockOSThread()")
	println("[main] Running on core:", machine.CurrentCore())
	runtime.UnlockOSThread()
	println("[main] Unpinned using runtime.UnlockOSThread()")
	println()

	// Example 2: Pin to a specific core using machine package
	machine.LockCore(0)
	println("[main] Explicitly pinned to core 0 using machine.LockCore()")
	println()

	// Start a goroutine pinned to core 1
	go core1Worker()

	// Start a goroutine using standard LockOSThread
	go standardLockWorker()

	// Start an unpinned goroutine (can run on either core)
	go unpinnedWorker()

	// Main loop on core 0
	for i := 0; i < 10; i++ {
		println("[main] loop", i, "on CPU", machine.CurrentCore())
		time.Sleep(500 * time.Millisecond)
	}

	// Unpin and let main run on any core
	machine.UnlockCore()
	println()
	println("[main] Unpinned using machine.UnlockCore()")

	// Continue running for a bit to show potential migration
	for i := 0; i < 5; i++ {
		println("[main] unpinned loop on CPU", machine.CurrentCore())
		time.Sleep(500 * time.Millisecond)
	}

	println()
	println("Example complete!")
}

// Worker function that pins to core 1 using explicit core selection
func core1Worker() {
	// Pin this goroutine to core 1 explicitly
	machine.LockCore(1)
	println("[core1-worker] Worker pinned to core 1 using machine.LockCore()")

	for i := 0; i < 10; i++ {
		println("[core1-worker] loop", i, "on CPU", machine.CurrentCore())
		time.Sleep(500 * time.Millisecond)
	}

	println("[core1-worker] Finished")
}

// Worker function that uses standard Go LockOSThread()
func standardLockWorker() {
	// Pin this goroutine to whichever core it starts on
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	core := machine.CurrentCore()
	println("[std-lock-worker] Worker locked using runtime.LockOSThread()")
	println("[std-lock-worker] Running on core:", core)

	for i := 0; i < 10; i++ {
		println("[std-lock-worker] loop", i, "on CPU", machine.CurrentCore())
		time.Sleep(600 * time.Millisecond)
	}

	println("[std-lock-worker] Finished")
}

// Worker function that is not pinned (can run on any core)
func unpinnedWorker() {
	println("[unpinned-worker] Starting")

	for i := 0; i < 10; i++ {
		cpu := machine.CurrentCore()
		println("[unpinned-worker] loop", i, "on CPU", cpu)
		time.Sleep(700 * time.Millisecond)

		// Yield to potentially migrate to another core
		runtime.Gosched()
	}

	println("[unpinned-worker] Finished")
}
