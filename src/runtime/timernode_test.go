package runtime_test

import (
	"runtime"
	"testing"
	"time"
)

func TestTimerResetAfterStopNoAlloc(t *testing.T) {
	tm := time.NewTimer(time.Hour)
	tm.Stop()
	allocs := testing.AllocsPerRun(100, func() {
		tm.Reset(time.Hour)
		tm.Stop()
	})
	if allocs != 0 {
		t.Errorf("Reset after Stop allocated %v times per run, want 0", allocs)
	}
}

func TestTimerNodesDroppedAfterGC(t *testing.T) {
	tm := time.NewTimer(time.Hour)
	tm.Stop()
	if n := runtime.TimerNodeFreeLen(); n == 0 {
		t.Fatal("stopped timer node was not kept for reuse")
	}
	runtime.GC()
	if n := runtime.TimerNodeFreeLen(); n != 0 {
		t.Errorf("free list holds %d nodes after a collection, want 0", n)
	}
}
