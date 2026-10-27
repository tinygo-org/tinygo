package main

import "runtime"

// Many rounds of leaked goroutines, interleaved with live goroutines that hold
// objects while parked. Freed stacks and tasks get reused, and the heap must
// not grow.

const (
	rounds  = 60
	leaked  = 200
	live    = 16
	warmup  = 10
	maxGrow = 4 << 20
)

type object struct {
	id   int
	data [14]int
}

//go:noinline
func newObject(id int) *object {
	o := &object{id: id}
	for i := range o.data {
		o.data[i] = id*31 + i
	}
	return o
}

func (o *object) check() {
	if o.id < 0 {
		panic("object was freed and reused")
	}
	for i, v := range o.data {
		if v != o.id*31+i {
			panic("object was corrupted")
		}
	}
}

func heapInuse() uint64 {
	runtime.GC()
	runtime.GC()
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	return m.HeapInuse
}

//go:noinline
func spawnLeaked(round int) {
	for i := 0; i < leaked; i++ {
		ch := make(chan int)
		o := newObject(round*1000 + i)
		go func() {
			<-ch
			o.check()
		}()
	}
}

//go:noinline
func spawnLive(round int, done chan struct{}) {
	for i := 0; i < live; i++ {
		go func(i int) {
			o := newObject(round*1000 + 500 + i)
			runtime.Gosched()
			runtime.GC()
			o.check()
			done <- struct{}{}
		}(i)
	}
}

func main() {
	var baseline uint64
	done := make(chan struct{})
	for round := 0; round < rounds; round++ {
		spawnLeaked(round)
		spawnLive(round, done)
		for i := 0; i < live; i++ {
			<-done
		}
		if round == warmup {
			baseline = heapInuse()
		}
	}
	final := heapInuse()
	if final > baseline+maxGrow {
		println("heap grew by", final-baseline, "bytes")
	} else {
		println("heap stayed flat")
	}
}
