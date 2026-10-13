// Assign goroutines to cores with machine.AssignCore.
// Build with -scheduler=cores.

//go:build scheduler.cores && (rp2040 || rp2350)

package main

import (
	"device/rp"
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
		println(name, "on core", currentCore())
		time.Sleep(700 * time.Millisecond)
	}
	println(name, "done")
}

// currentCore reads the core number, which is stable while assigned.
func currentCore() uint32 {
	return rp.SIO.CPUID.Get()
}
