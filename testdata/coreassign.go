package main

import (
	"machine"
	"runtime"
	"sync"
	"sync/atomic"
	"time"
	_ "unsafe"
)

//go:linkname currentCPU runtime.currentCPU
func currentCPU() uint32

func main() {
	numCPU := runtime.NumCPU()

	// Assign one goroutine to each core while unassigned goroutines keep the cores busy.
	var wg sync.WaitGroup
	done := make(chan struct{})
	for i := 0; i < numCPU; i++ {
		go func() {
			for {
				select {
				case <-done:
					return
				default:
					runtime.Gosched()
				}
			}
		}()
	}
	results := make([]bool, numCPU)
	for core := 0; core < numCPU; core++ {
		wg.Add(1)
		go func(core int) {
			defer wg.Done()
			machine.AssignCore(core)
			ok := currentCPU() == uint32(core)
			for i := 0; i < 50; i++ {
				if i%2 == 0 {
					runtime.Gosched()
				} else {
					time.Sleep(time.Millisecond)
				}
				ok = ok && currentCPU() == uint32(core)
			}
			results[core] = ok
		}(core)
	}
	wg.Wait()
	close(done)
	for core, ok := range results {
		println("AssignCore", core, "stays on core:", ok)
	}

	// A second AssignCore panics.
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer func() {
			println("second AssignCore panics:", recover() != nil)
		}()
		machine.AssignCore(0)
		machine.AssignCore(0)
	}()
	wg.Wait()

	// Move main to the last core.
	target := numCPU - 1
	machine.AssignCore(target)
	println("AssignCore moves main:", currentCPU() == uint32(target))

	// An assigned task woken from another core resumes on its own core.
	ch := make(chan int)
	ack := make(chan bool)
	var mu sync.Mutex
	workerOK := make(chan bool)
	go func() {
		machine.AssignCore(0)
		ok := currentCPU() == 0
		for i := 0; i < 20; i++ {
			mu.Lock()
			ch <- i
			ok = ok && currentCPU() == 0
			time.Sleep(time.Millisecond)
			ok = ok && currentCPU() == 0
			mu.Unlock()
			<-ack
			ok = ok && currentCPU() == 0
		}
		workerOK <- ok
	}()
	woken := true
	for i := 0; i < 20; i++ {
		<-ch
		ok := currentCPU() == uint32(target)
		mu.Lock()
		ok = ok && currentCPU() == uint32(target)
		mu.Unlock()
		woken = woken && ok
		ack <- ok
	}
	println("assigned task woken from another core:", woken)
	println("assigned waker stays on core:", <-workerOK)

	// A task sleeping on one core wakes on time while a task on another core is busy.
	var busyOver atomic.Bool
	wokeInTime := make(chan bool)
	go func() {
		machine.AssignCore(0)
		time.Sleep(20 * time.Millisecond)
		end := time.Now().Add(300 * time.Millisecond)
		for time.Now().Before(end) {
		}
		busyOver.Store(true)
	}()
	go func() {
		machine.AssignCore(1)
		time.Sleep(50 * time.Millisecond)
		wokeInTime <- !busyOver.Load()
	}()
	println("sleeping task wakes while another core is busy:", <-wokeInTime)
}
