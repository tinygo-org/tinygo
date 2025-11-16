// This example demonstrates goroutine core pinning on multi-core systems (RP2040/RP2350).
// It shows how to pin goroutines to specific CPU cores and verify their execution.
//

//go:build rp2040 || rp2350

package main

import (
	"runtime"
	"time"
)

func main() {
	println("=== Core Pinning Example ===")
	println("Number of CPU cores:", runtime.NumCPU())
	println("Main starting on core:", runtime.CurrentCPU())
	println()

	// Pin main goroutine to core 0
	runtime.LockToCore(0)
	println("Main pinned to core:", runtime.GetAffinity())
	println()

	// Start a goroutine pinned to core 1
	go core1Worker()

	// Start an unpinned goroutine (can run on either core)
	go unpinnedWorker()

	// Main loop on core 0
	for i := 0; i < 10; i++ {
		println("Core 0 (main):", i, "on CPU", runtime.CurrentCPU())
		time.Sleep(500 * time.Millisecond)
	}

	// Unpin and let main run on any core
	runtime.UnlockFromCore()
	println()
	println("Main unpinned, affinity:", runtime.GetAffinity())

	// Continue running for a bit to show migration
	for i := 0; i < 5; i++ {
		println("Unpinned main on CPU", runtime.CurrentCPU())
		time.Sleep(500 * time.Millisecond)
	}

	println()
	println("Example complete!")
}

// Worker function that runs on core 1
func core1Worker() {
	// Pin this goroutine to core 1
	runtime.LockToCore(1)
	println("Worker pinned to core:", runtime.GetAffinity())

	for i := 0; i < 10; i++ {
		println("  Core 1 (worker):", i, "on CPU", runtime.CurrentCPU())
		time.Sleep(500 * time.Millisecond)
	}

	println("  Core 1 worker finished")
}

// Worker function that is not pinned (can run on any core)
func unpinnedWorker() {
	println("Unpinned worker starting, affinity:", runtime.GetAffinity())

	for i := 0; i < 10; i++ {
		cpu := runtime.CurrentCPU()
		println("    Unpinned worker:", i, "on CPU", cpu)
		time.Sleep(700 * time.Millisecond)

		// Yield to potentially migrate to another core
		runtime.Gosched()
	}

	println("    Unpinned worker finished")
}
