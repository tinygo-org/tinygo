//go:build tinygo && !scheduler.none

package main

import (
	"strconv"
	"sync"
	"time"

	"github.com/tinygo-org/tinygo/tools/qbs/dev"
)

var sink []byte

func init() {
	extraTests = append(extraTests, dev.Test{Name: "conc", Run: conc})
}

func worker(id int, wg *sync.WaitGroup) {
	defer wg.Done()
	for k := 0; k < 200; k++ {
		sink = make([]byte, 256)
		println("G", id, k)
		time.Sleep(time.Millisecond)
	}
}

// conc prints from goroutines while the host sends bytes.
func conc(t *dev.T) {
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
		for dev.Serial.Buffered() > 0 {
			dev.Serial.ReadByte()
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
	for dev.Serial.Buffered() > 0 {
		dev.Serial.ReadByte()
		rx++
	}
	t.Val("rx", strconv.Itoa(rx))
}
