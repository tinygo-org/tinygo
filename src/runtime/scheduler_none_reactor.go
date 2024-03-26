//go:build scheduler.none

package runtime

// runReactor is the program entry point for a WebAssembly reactor program, instead of run().
// With the "none" scheduler, init (but not main) functions are invoked directly.
func runReactor() {
	initHeap()
	initAll()
	// main is NOT called
}
