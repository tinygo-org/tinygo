//go:build (gc.conservative || gc.precise) && !scheduler.none && !scheduler.tasks && !scheduler.asyncify

package runtime

// spawnFinalizerRunner is defined once per scheduler class, and the three build
// constraints partition the scheduler space exactly (exactly one scheduler.* tag
// is ever set): scheduler.none in gc_finalizer_sched_none.go, scheduler.tasks and
// scheduler.asyncify in gc_finalizer_sched.go, and every other variant here. This
// is the catch-all, so a new scheduler variant lands here and stays defined
// rather than falling through to an undefined reference.
//
// Non-cooperative schedulers (cores, threads) spawn the finalizer runner but have
// no cooperative idle point, so they do not install the idle-pressure collector;
// the runner drains finalizers as GCs queue them.
func spawnFinalizerRunner() { go finalizerRunner() }
