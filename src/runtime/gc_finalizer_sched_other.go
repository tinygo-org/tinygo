//go:build (gc.conservative || gc.precise) && !scheduler.none && !scheduler.tasks && !scheduler.asyncify

package runtime

// spawnFinalizerRunner is the fallback for noncooperative schedulers.
// These schedulers run finalizers but do not install the idle GC hook.
func spawnFinalizerRunner() { go finalizerRunner() }
