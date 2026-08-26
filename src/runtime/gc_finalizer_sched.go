//go:build (gc.conservative || gc.precise) && (scheduler.tasks || scheduler.asyncify)

package runtime

// Keep this setup in a file for these schedulers so unused finalizer code can be removed.
// Cooperative schedulers also install the idle GC hook.
func spawnFinalizerRunner() {
	finalizerIdleGC = finalizerPressureGC
	go finalizerRunner()
}
