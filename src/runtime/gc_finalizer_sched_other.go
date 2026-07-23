//go:build (gc.conservative || gc.precise) && !scheduler.none && !scheduler.tasks && !scheduler.asyncify

package runtime

// Non-cooperative schedulers (cores, threads) spawn the finalizer runner but
// have no cooperative idle point, so they do not install the idle-pressure
// collector; the runner drains finalizers as GCs queue them.
func spawnFinalizerRunner() { go finalizerRunner() }
