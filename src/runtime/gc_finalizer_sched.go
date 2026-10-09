//go:build (gc.conservative || gc.precise || gc.boehm) && (scheduler.tasks || scheduler.asyncify)

package runtime

// Keep this setup in a file for these schedulers so unused finalizer code can be removed.
// Cooperative schedulers also install the idle GC hook.
func spawnFinalizerRunner() {
	initFinalizerScheduler()
	go finalizerRunner()
}

func initFinalizerScheduler() { finalizerIdleGC = finalizerPressureGC }
