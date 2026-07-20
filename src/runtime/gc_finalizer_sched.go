//go:build (gc.conservative || gc.precise) && !scheduler.none

package runtime

// The go statement lives in this scheduler-gated file, not inline in
// registerFinalizer, so scheduler.none builds never reference internal/task.start
// and the runner is DCE'd when SetFinalizer is unused.
func spawnFinalizerRunner() { go finalizerRunner() }
