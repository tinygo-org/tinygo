//go:build scheduler.threads

package runtime_test

import (
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
