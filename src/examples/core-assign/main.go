// Assign goroutines to cores with machine.AssignCore.
// Build with -scheduler=cores.

//go:build scheduler.cores && (rp2040 || rp2350 || esp32s3)

package main

import (
	"machine"
	"runtime"
	"time"
)

func main() {
	time.Sleep(2 * time.Second)
	println("cores:", runtime.NumCPU())

	go worker("core1 worker", 1)
	go worker("unassigned worker", -1)
	go worker("core0 worker", 0)

	for {
		time.Sleep(time.Second)
	}
}

func worker(name string, core int) {
	if core >= 0 {
		machine.AssignCore(core)
	}
	for i := 0; i < 10; i++ {
		println(name, "on core", machine.CurrentCore())
		time.Sleep(700 * time.Millisecond)
	}
	println(name, "done")
}
