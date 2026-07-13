//go:build (gc.conservative || gc.precise) && scheduler.none

package runtime

// scheduler.none has no goroutines; finalizers drain inline in wakeFinalizer.
func spawnFinalizerRunner() {}
