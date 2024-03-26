package main

import (
	"time"
	_ "unsafe"
)

//go:linkname scheduler runtime.scheduler
func scheduler()

//go:linkname setSchedulerDone runtime.setSchedulerDone
func setSchedulerDone(bool)

// __go_wasm_export_tinygo_test is a wrapper function around tinygo_test
// that runs the exported function in a goroutine and starts the scheduler.
// Goroutines started by this or other functions will persist, are paused
// when this function returns, and restarted when the host calls back into
// another exported function.
//
//export tinygo_test
func __go_wasm_export_tinygo_test() int32 {
	setSchedulerDone(false)
	var ret int32
	go func() {
		ret = tinygo_test()
		setSchedulerDone(true)
	}()
	scheduler()
	return ret
}

func tinygo_test() int32 {
	for ticks != 1337 {
		time.Sleep(time.Nanosecond)
	}
	return ticks
}

var ticks int32

func init() {
	// Start infinite ticker
	go func() {
		for {
			ticks++
			time.Sleep(time.Nanosecond)
		}
	}()
}
