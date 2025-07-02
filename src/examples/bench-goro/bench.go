package main

import (
	"machine"
	"runtime"
	"sync"
	"time"
)

const N = 500000
const Ngoro = 4

func main() {
	start := time.Now()
	var wg sync.WaitGroup
	wg.Add(Ngoro)
	for i := 0; i < Ngoro; i++ {
		go adder(&wg, N)
	}
	wg.Wait()
	elapsed := time.Since(start)
	goroutineCtxSwitchOverhead := (elapsed / (Ngoro * N)).String()

	elapsedstr := elapsed.String()
	machine.LED.Configure(machine.PinConfig{Mode: machine.PinOutput})
	for {
		println("bench:", elapsedstr, "goroutine ctx switch:", goroutineCtxSwitchOverhead)
		machine.LED.High()
		time.Sleep(elapsed)
		machine.LED.Low()
		time.Sleep(elapsed)
	}
}

func adder(wg *sync.WaitGroup, num int) {
	for i := 0; i < num; i++ {
		runtime.Gosched()
	}
	wg.Done()
}
