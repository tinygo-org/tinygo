//go:build scheduler.threads && (darwin || (linux && !baremetal && !wasm_unknown && !wasip2 && !nintendoswitch))

package runtime_test

import (
	"runtime"
	"testing"
	"time"
)

// Under the threads scheduler, sleeping must not consume signalFutex. That
// belongs to the signal watcher.
func TestSleepTicksLeavesSignalFutex(t *testing.T) {
	const d = 20 * time.Millisecond
	runtime.SetSignalFutex(1)
	defer runtime.SetSignalFutex(0)

	start := time.Now()
	runtime.SleepTicksNs(int64(d))
	if elapsed := time.Since(start); elapsed < d {
		t.Errorf("slept %v, want at least %v", elapsed, d)
	}
	if got := runtime.GetSignalFutex(); got != 1 {
		t.Errorf("signalFutex = %d after sleep, want 1", got)
	}
}
