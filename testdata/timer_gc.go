package main

/*
#cgo CFLAGS: -I${SRCDIR}/../lib/bdwgc/include
#include <gc/gc_mark.h>

static void count_object(void *object, size_t bytes, void *data) {
    (*(size_t *)data)++;
}

static void *count_locked(void *data) {
    GC_enumerate_reachable_objects_inner(count_object, data);
    return NULL;
}

static size_t live_objects(void) {
    size_t count = 0;
    GC_call_with_alloc_lock(count_locked, &count);
    return count;
}
*/
import "C"

import (
	"runtime"
	"time"
	_ "unsafe"
)

func collectTimers() int64 {
	runtime.GC()
	runtime.GC()
	runtime.GC()
	return int64(C.live_objects())
}

func main() {
	testRealTimerInBubble()
	ready := make(chan struct{})
	close(ready)
	operations := []func(){
		func() { time.After(time.Hour) },
		func() { time.Tick(time.Hour) },
		func() { time.NewTimer(time.Hour) },
		func() { time.NewTicker(time.Hour) },
		func() { time.NewTimer(time.Hour).Stop() },
		func() { time.NewTicker(time.Hour).Stop() },
		func() { time.NewTimer(time.Hour).Reset(time.Hour) },
		func() {
			select {
			case <-ready:
			case <-time.After(time.Hour):
				panic("future timer fired")
			}
		},
		func() {
			gate := make(chan struct{})
			go func() { close(gate) }()
			select {
			case <-gate:
			case <-time.After(time.Hour):
				panic("future timer fired")
			}
		},
	}
	for _, operation := range operations {
		for i := 0; i < 16; i++ {
			operation()
		}
		before := collectTimers()
		for i := 0; i < 10000; i++ {
			operation()
		}
		if delta := collectTimers() - before; delta >= 64 {
			println("retained timer objects:", delta)
			panic("discarded timer remained reachable")
		}
	}
	println("timer collection tests passed")
}

func testRealTimerInBubble() {
	timer := time.NewTimer(time.Millisecond)
	synctestRun(func() {
		<-timer.C
	})
	if timer.Stop() {
		panic("bubble returned before real timer fired")
	}
	before := time.Now().Add(-time.Second)
	timer = time.NewTimer(-time.Nanosecond)
	synctestRun(func() {
		if value := <-timer.C; value.Before(before) {
			panic("real timer used bubble clock")
		}
	})
}

//go:linkname synctestRun internal/synctest.Run
func synctestRun(func())
