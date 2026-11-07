// Timers are stopped and reset by goroutines that can run on other cores or
// threads while the timers fire. Every timer activation must either fire or be
// cancelled by a call that returned true, never both and never neither.
package main

import (
	"sync"
	"sync/atomic"
	"time"
)

const (
	timers = 300
	rounds = 8
)

type rng uint32

func (r *rng) next() uint32 {
	x := uint32(*r)
	x ^= x << 13
	x ^= x >> 17
	x ^= x << 5
	*r = rng(x)
	return x
}

func spin(n uint32) {
	for i := uint32(0); i < n; i++ {
		_ = i * i
	}
}

var failed atomic.Bool

func fail(msg string, a, b int) {
	failed.Store(true)
	println("FAIL", msg, a, b)
}

// stopRace: each timer fires once or is stopped, exactly one of the two.
func stopRace(seed uint32) {
	fires := make([]atomic.Int32, timers)
	stopped := make([]bool, timers)
	tm := make([]*time.Timer, timers)
	for i := range tm {
		i := i
		tm[i] = time.AfterFunc(time.Duration(i%4)*100*time.Microsecond, func() { fires[i].Add(1) })
	}
	var wg sync.WaitGroup
	for g := 0; g < 3; g++ {
		wg.Add(1)
		r := rng(seed + uint32(g)*7919 + 1)
		go func(g int) {
			defer wg.Done()
			for i := g; i < timers; i += 3 {
				spin(r.next() % 4000)
				stopped[i] = tm[i].Stop()
			}
		}(g)
	}
	wg.Wait()
	time.Sleep(200 * time.Millisecond)
	for i := range tm {
		f := int(fires[i].Load())
		s := 0
		if stopped[i] {
			s = 1
		}
		if f+s != 1 {
			fail("stop: fires+stopped", f, s)
		}
	}
}

// resetRace: fires must equal 1 plus the number of Reset calls that returned false.
func resetRace(seed uint32) {
	fires := make([]atomic.Int32, timers)
	falseResets := make([]int, timers)
	tm := make([]*time.Timer, timers)
	for i := range tm {
		i := i
		tm[i] = time.AfterFunc(time.Duration(i%4)*100*time.Microsecond, func() { fires[i].Add(1) })
	}
	var wg sync.WaitGroup
	for g := 0; g < 3; g++ {
		wg.Add(1)
		r := rng(seed + uint32(g)*104729 + 1)
		go func(g int) {
			defer wg.Done()
			for i := g; i < timers; i += 3 {
				for k := 0; k < 4; k++ {
					spin(r.next() % 4000)
					if !tm[i].Reset(time.Duration(r.next()%3) * 100 * time.Microsecond) {
						falseResets[i]++
					}
				}
			}
		}(g)
	}
	wg.Wait()
	time.Sleep(200 * time.Millisecond)
	for i := range tm {
		f := int(fires[i].Load())
		if f != 1+falseResets[i] {
			fail("reset: fires vs 1+false", f, 1+falseResets[i])
		}
	}
}

// tickerStop: a stopped ticker must stay stopped even if Stop races with its tick.
func tickerStop(seed uint32) {
	const n = 40
	var ticks [n]atomic.Int32
	tk := make([]*time.Ticker, n)
	for i := range tk {
		i := i
		tk[i] = time.NewTicker(200 * time.Microsecond)
		go func() {
			for range tk[i].C {
				ticks[i].Add(1)
			}
		}()
	}
	var wg sync.WaitGroup
	for g := 0; g < 2; g++ {
		wg.Add(1)
		r := rng(seed + uint32(g)*31 + 1)
		go func(g int) {
			defer wg.Done()
			for i := g; i < n; i += 2 {
				spin(r.next() % 20000)
				tk[i].Stop()
			}
		}(g)
	}
	wg.Wait()
	time.Sleep(5 * time.Millisecond)
	var before [n]int32
	for i := range before {
		before[i] = ticks[i].Load()
	}
	time.Sleep(200 * time.Millisecond)
	for i := range before {
		if after := ticks[i].Load(); after != before[i] {
			fail("ticker kept ticking after Stop", int(before[i]), int(after))
		}
	}
}

func main() {
	for round := 0; round < rounds; round++ {
		seed := uint32(round)*2654435761 + 12345
		stopRace(seed)
		resetRace(seed)
		tickerStop(seed)
		if failed.Load() {
			println("FAILED at round", round)
			return
		}
	}
	println("timer stop reset stress ok")
}
