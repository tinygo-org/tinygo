//go:build scheduler.threads

package runtime

import "internal/task"

func gcMarkReachable() {
	task.GCStopWorldAndScan()
}

// Scan globals inside the stop-the-world phase. Called from the STW
// implementation in the internal/task package.
//
//go:linkname gcScanGlobals internal/task.gcScanGlobals
func gcScanGlobals() {
	findGlobals(markRoots)
}

// Function called from assembly with all registers pushed, to actually scan the
// stack.
//
//go:export tinygo_scanstack
func scanstack(sp uintptr) {
	markRoots(sp, task.StackTop())
}

func gcResumeWorld() {
	task.GCResumeWorld()
}
