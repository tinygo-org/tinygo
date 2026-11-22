// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package testing

import (
	"internal/synctest"
	"reflect"
	"strings"
	"time"
)

func TestSynctestFailureOutput(t *T) {
	parent := &T{common: common{output: &logger{}, name: "parent"}}
	parent.Log("before bubble")
	ok := testingSynctestTest(parent, func(t *T) {
		t.Error("inside bubble")
	})
	parent.Log("after bubble")

	if ok || !parent.Failed() {
		t.Fatalf("bubble success = %v, parent failed = %v", ok, parent.Failed())
	}
	output := parent.output.b.String()
	before := strings.Index(output, "before bubble")
	inside := strings.Index(output, "inside bubble")
	after := strings.Index(output, "after bubble")
	if strings.Contains(output, "--- FAIL:") ||
		before < 0 || inside <= before || after <= inside ||
		strings.Count(output, "inside bubble") != 1 {
		t.Errorf("unexpected parent log:\n%s", output)
	}
}

func TestSynctestBubbleSkip(t *T) {
	if !supportsRecover() {
		t.Skip("requires runtime.Goexit")
		return
	}
	parent := &T{common: common{output: &logger{}, name: "parent"}}
	ok := testingSynctestTest(parent, func(t *T) {
		t.Skip("skip bubble")
	})

	if !ok || parent.Failed() || parent.Skipped() {
		t.Fatalf("bubble success = %v, parent failed = %v, parent skipped = %v",
			ok, parent.Failed(), parent.Skipped())
	}
	output := parent.output.b.String()
	if !strings.Contains(output, "skip bubble") || strings.Contains(output, "--- ") {
		t.Errorf("unexpected parent log:\n%s", output)
	}
}

func TestSynctestDuringCleanup(t *T) {
	parent := &T{}
	parent.cleanupStarted.Store(true)
	defer func() {
		const want = "testing: synctest.Run called during t.Cleanup"
		if got := recover(); got != want {
			t.Errorf("panic = %v, want %q", got, want)
		}
	}()
	testingSynctestTest(parent, func(*T) {})
}

func TestSynctestAcquireDelaysRun(t *T) {
	acquired := make(chan *synctest.Bubble)
	release := make(chan struct{})
	runDone := make(chan struct{})

	go func() {
		bubble := <-acquired
		<-release
		bubble.Release()
	}()

	go func() {
		synctest.Run(func() {
			acquired <- synctest.Acquire()
		})
		close(runDone)
	}()

	select {
	case <-runDone:
		t.Fatal("synctest.Run returned before the bubble reference was released")
	case <-time.After(time.Millisecond):
	}
	close(release)
	<-runDone
}

func TestSynctestSleepOverflow(t *T) {
	synctest.Run(func() {
		start := time.Now()
		time.Sleep(time.Duration(1<<63 - 1))
		if elapsed := time.Since(start); elapsed == 0 {
			t.Fatal("maximum-duration sleep returned without advancing fake time")
		}
		time.Sleep(time.Nanosecond)
		synctest.Wait()
	})
}

func TestSynctestTickerConcurrentStopAndReset(t *T) {
	for range 100 {
		synctest.Run(func() {
			ticker := time.NewTicker(time.Nanosecond)
			stopped := make(chan struct{})
			go func() {
				<-ticker.C
				ticker.Stop()
				close(stopped)
			}()
			<-stopped
			time.Sleep(time.Nanosecond)
			select {
			case <-ticker.C:
				t.Fatal("stopped ticker fired again")
			default:
			}
		})

		synctest.Run(func() {
			ticker := time.NewTicker(time.Nanosecond)
			reset := make(chan struct{})
			go func() {
				<-ticker.C
				ticker.Reset(10 * time.Nanosecond)
				close(reset)
			}()
			<-reset
			start := time.Now()
			<-ticker.C
			if elapsed := time.Since(start); elapsed != 10*time.Nanosecond {
				t.Fatalf("reset ticker fired after %v, want 10ns", elapsed)
			}
			ticker.Stop()
		})
	}
}

func TestSynctestTimerImmediateReset(t *T) {
	synctest.Run(func() {
		timer := time.NewTimer(time.Hour)
		timer.Reset(0)
		<-timer.C
	})
}

func TestSynctestTickerConcurrentResets(t *T) {
	for range 100 {
		synctest.Run(func() {
			ticker := time.NewTicker(time.Hour)
			start := make(chan struct{})
			done := make(chan struct{}, 2)
			for _, duration := range []time.Duration{10, 20} {
				go func() {
					<-start
					ticker.Reset(duration)
					done <- struct{}{}
				}()
			}
			close(start)
			<-done
			<-done
			ticker.Stop()
			time.Sleep(100 * time.Nanosecond)
			select {
			case <-ticker.C:
				t.Fatal("stopped ticker fired after concurrent resets")
			default:
			}
		})
	}
}

func TestCleanup(t *T) {
	var cleanups []int
	t.Run("test", func(t *T) {
		t.Cleanup(func() { cleanups = append(cleanups, 1) })
		t.Cleanup(func() { cleanups = append(cleanups, 2) })
	})
	if got, want := cleanups, []int{2, 1}; !reflect.DeepEqual(got, want) {
		t.Errorf("unexpected cleanup record; got %v want %v", got, want)
	}
}

func TestRunCleanup(t *T) {
	outerCleanup := 0
	innerCleanup := 0
	t.Run("test", func(t *T) {
		t.Cleanup(func() { outerCleanup++ })
		t.Run("x", func(t *T) {
			t.Cleanup(func() { innerCleanup++ })
		})
	})
	if innerCleanup != 1 {
		t.Errorf("unexpected inner cleanup count; got %d want 1", innerCleanup)
	}
	if outerCleanup != 1 {
		t.Errorf("unexpected outer cleanup count; got %d want 1", outerCleanup) // wrong upstream!
	}
}

func TestCleanupParallelSubtests(t *T) {
	ranCleanup := 0
	t.Run("test", func(t *T) {
		t.Cleanup(func() { ranCleanup++ })
		t.Run("x", func(t *T) {
			t.Parallel()
			if ranCleanup > 0 {
				t.Error("outer cleanup ran before parallel subtest")
			}
		})
	})
	if ranCleanup != 1 {
		t.Errorf("unexpected cleanup count; got %d want 1", ranCleanup)
	}
}

func TestNestedCleanup(t *T) {
	ranCleanup := 0
	t.Run("test", func(t *T) {
		t.Cleanup(func() {
			if ranCleanup != 2 {
				t.Errorf("unexpected cleanup count in first cleanup: got %d want 2", ranCleanup)
			}
			ranCleanup++
		})
		t.Cleanup(func() {
			if ranCleanup != 0 {
				t.Errorf("unexpected cleanup count in second cleanup: got %d want 0", ranCleanup)
			}
			ranCleanup++
			t.Cleanup(func() {
				if ranCleanup != 1 {
					t.Errorf("unexpected cleanup count in nested cleanup: got %d want 1", ranCleanup)
				}
				ranCleanup++
			})
		})
	})
	if ranCleanup != 3 {
		t.Errorf("unexpected cleanup count: got %d want 3", ranCleanup)
	}
}
