//go:build (gc.conservative || gc.precise || gc.boehm) && scheduler.none

package runtime

// scheduler.none has no goroutines; finalizers drain inline in wakeFinalizer.
func spawnFinalizerRunner() {}
