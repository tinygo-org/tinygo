//go:build !scheduler.none

package runtime

// runReactor is the program entry point for a WebAssembly reactor program, instead of run().
// With a scheduler, init (but not main) functions are invoked in a goroutine before starting the scheduler.
func runReactor() {
	initHeap()
	go func() {
		initAll()
		// main is NOT called
		schedulerDone = true
	}()
	scheduler()
}
