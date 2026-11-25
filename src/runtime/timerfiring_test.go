package runtime_test

import (
	"runtime"
	"testing"
	"time"
	"unsafe"
)

func TestTimerStopTwiceWhileFiring(t *testing.T) {
	if !runtime.TimersFireConcurrently {
		t.Skip("timer callbacks do not run concurrently with goroutines")
	}
	tm := time.NewTimer(time.Hour)
	p := unsafe.Pointer(tm)
	tn := runtime.TimerBeginFiring(p)
	if !tm.Stop() {
		t.Error("first Stop returned false for a firing timer")
	}
	if tm.Stop() {
		t.Error("second Stop returned true for a stopped timer")
	}
	if tm.Reset(time.Hour) {
		t.Error("Reset returned true for a stopped timer")
	}
	runtime.TimerFinishFiring(p, tn)
	tm.Stop()
}
