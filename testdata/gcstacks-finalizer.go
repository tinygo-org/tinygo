package main

import (
	"runtime"
	"time"
)

// A goroutine blocks on a channel that only a finalizable object references.
// Resurrecting that object must keep the goroutine and its locals alive.

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

func (o *object) check(what string) {
	if o.id < 0 {
		panic(what + ": object was freed and reused")
	}
	for i, v := range o.data {
		if v != o.id*31+i {
			panic(what + ": object was corrupted")
		}
	}
}

func churn() {
	var all [][]*object
	for i := 0; i < 64; i++ {
		var batch []*object
		for j := 0; j < 16; j++ {
			batch = append(batch, newObject(-1-j))
		}
		all = append(all, batch)
	}
	runtime.KeepAlive(all)
}

func collect() {
	runtime.GC()
	churn()
	runtime.GC()
	churn()
}

type holder struct{ c chan int }

var (
	fired  int
	result = make(chan int)
)

func worker(c chan int) {
	a := newObject(1)
	b := newObject(2)
	v := <-c
	collect()
	a.check("worker a")
	b.check("worker b")
	result <- v + a.id + b.id
}

//go:noinline
func setup() {
	h := &holder{c: make(chan int)}
	runtime.SetFinalizer(h, func(h *holder) {
		fired++
		go func() { h.c <- 7 }()
	})
	go worker(h.c)
}

// scrubStack overwrites stale words left by setup.
//
//go:noinline
func scrubStack(depth int) int {
	if depth <= 0 {
		return 0
	}
	var buf [64]int
	for i := range buf {
		buf[i] = depth + i
	}
	return scrubStack(depth-1) + buf[depth&63]
}

var sink int

func main() {
	setup()
	sink += scrubStack(8)
	runtime.Gosched()
	for i := 0; i < 200 && fired == 0; i++ {
		collect()
		runtime.Gosched()
		time.Sleep(time.Millisecond)
	}
	if fired == 0 {
		println("finalizer did not run")
		return
	}
	println("finalizer ran, worker result:", <-result)
}
