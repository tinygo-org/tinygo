//go:build tinygo

package main

import (
	"sync"
	"time"

	"machine"

	"github.com/tinygo-org/tinygo/tools/uart-test/common"
)

var sink []byte

func worker(id int, wg *sync.WaitGroup) {
	defer wg.Done()
	for k := 0; k < 200; k++ {
		sink = make([]byte, 256)
		println("G", id, k)
		time.Sleep(time.Millisecond)
	}
}

func main() {
	common.WaitGo("conc")
	println("@@GO conc")

	var wg sync.WaitGroup
	for id := 0; id < 3; id++ {
		wg.Add(1)
		go worker(id, &wg)
	}
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	rx := 0
	finished := false
	for !finished {
		for machine.Serial.Buffered() > 0 {
			machine.Serial.ReadByte()
			rx++
		}
		select {
		case <-done:
			finished = true
		default:
			time.Sleep(time.Millisecond)
		}
	}
	// Let the host finish sending, then count what is left.
	time.Sleep(500 * time.Millisecond)
	for machine.Serial.Buffered() > 0 {
		machine.Serial.ReadByte()
		rx++
	}
	println("@@RXCOUNT", rx)
	println("@@END conc")
	for {
		time.Sleep(time.Second)
	}
}
