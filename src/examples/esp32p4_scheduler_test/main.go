// esp32p4_scheduler_test exercises Phase 2 deliverables on the ESP32-P4:
//   - Multiple goroutines running concurrently via the TIMG0 alarm scheduler
//   - FPU callee-saved register correctness across goroutine switches
//   - Channel communication between goroutines
//   - time.Sleep and time.Since() accuracy
//
// Expected serial output (one block per second, indefinitely):
//
//	[tick 1] elapsed: ~1000ms
//	  float A=1.100 B=2.200 (each goroutine accumulates independently)
//	  chan: got 42
//	[tick 2] elapsed: ~1000ms
//	  float A=2.200 B=4.400
//	  chan: got 42
//	...
//
// If the FPU context save is broken, goroutine A and B will see each
// other's accumulator values and the floats will diverge unexpectedly.
package main

import (
	"time"
)

func main() {
	ch := make(chan int, 1)

	// Goroutine A: accumulates by 1.1 each second.
	go floatWorker("A", 1.1, ch)

	// Goroutine B: accumulates by 2.2 each second, sends on channel.
	go floatWorker("B", 2.2, ch)

	start := time.Now()
	tick := 0
	for {
		time.Sleep(time.Second)
		tick++
		elapsed := time.Since(start)
		ms := elapsed.Milliseconds()
		println("[tick", tick, "] elapsed:", ms, "ms")

		// Drain whatever the B goroutine sent.
		select {
		case v := <-ch:
			println("  chan: got", v)
		default:
			println("  chan: nothing yet")
		}
	}
}

// floatWorker accumulates a float32 by delta each second.
// It holds the accumulator in a local variable across time.Sleep calls,
// exercising FPU callee-saved register save/restore in the context switch.
func floatWorker(name string, delta float32, ch chan<- int) {
	var acc float32
	for {
		time.Sleep(time.Second)
		acc += delta
		// Use println to print the float (avoids strconv/fmt float path).
		print("  float ")
		print(name)
		print("=")
		println(acc)
		// B sends a sentinel to verify channel delivery.
		if name == "B" {
			select {
			case ch <- 42:
			default:
			}
		}
	}
}
