//go:build (gc.conservative || gc.precise) && (scheduler.tasks || scheduler.asyncify)

package runtime

// The go statement and the idle-hook install live in this scheduler-gated file,
// not inline in registerFinalizer, so a build that never calls SetFinalizer
// keeps internal/task.start and the whole finalizer collection path DCE'd. The
// cooperative scheduler additionally collects on finalizer-registration pressure
// at its idle point (see finalizerIdleGC in scheduler_cooperative.go).
func spawnFinalizerRunner() {
	finalizerIdleGC = finalizerPressureGC
	go finalizerRunner()
}
