//go:build tinygo

package main

import (
	"runtime/volatile"
	"strconv"
	"time"

	"github.com/tinygo-org/tinygo/tools/qbs/dev"
)

func main() {
	dev.Main("timing", []dev.Test{
		{Name: "sleep-100ms", Run: sleep100},
		{Name: "sleep-1ms", Run: sleep1},
		{Name: "monotonic", Run: monotonic},
		{Name: "busy-loop", Run: busyLoop},
		{Name: "clock-rate", Run: clockRate},
	})
}

func sleep100(t *dev.T) {
	var min, max time.Duration
	for i := 0; i < 20; i++ {
		start := time.Now()
		time.Sleep(100 * time.Millisecond)
		d := time.Since(start)
		if i == 0 || d < min {
			min = d
		}
		if d > max {
			max = d
		}
	}
	t.Val("min_us", strconv.Itoa(int(min.Microseconds())))
	t.Val("max_us", strconv.Itoa(int(max.Microseconds())))
	if min < 95*time.Millisecond {
		t.Fail("a 100ms sleep returned after " + min.String())
	}
	if max > 200*time.Millisecond {
		t.Fail("a 100ms sleep took " + max.String())
	}
}

func sleep1(t *dev.T) {
	start := time.Now()
	for i := 0; i < 100; i++ {
		time.Sleep(time.Millisecond)
	}
	d := time.Since(start)
	t.Val("total_us", strconv.Itoa(int(d.Microseconds())))
	if d < 90*time.Millisecond {
		t.Fail("100 sleeps of 1ms took only " + d.String())
	}
	if d > time.Second {
		t.Fail("100 sleeps of 1ms took " + d.String())
	}
}

func monotonic(t *dev.T) {
	start := time.Now()
	prev := start
	calls := 0
	for time.Since(start) < 50*time.Millisecond {
		now := time.Now()
		if now.Sub(prev) < 0 {
			t.Fail("time.Now went backwards")
			return
		}
		prev = now
		calls++
	}
	t.Val("calls_per_ms", strconv.Itoa(calls/50))
}

func busyLoop(t *dev.T) {
	start := time.Now()
	// The volatile variable keeps the compiler from removing the loop or
	// moving it past the clock reads.
	var v volatile.Register32
	v.Set(2463534242)
	for i := 0; i < 100000; i++ {
		x := v.Get()
		x ^= x << 13
		x ^= x >> 17
		x ^= x << 5
		v.Set(x)
	}
	t.Val("us_per_100k_iterations", strconv.Itoa(int(time.Since(start).Microseconds())))
}

// clockRate prints the device clock so the host can compare it with its own.
func clockRate(t *dev.T) {
	start := time.Now()
	for i := 0; i < 30; i++ {
		time.Sleep(100 * time.Millisecond)
		println("@@TICK", i, time.Since(start).Microseconds())
	}
}
