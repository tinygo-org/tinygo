package main

import (
	"runtime"
	"sync"
	"time"
)

// Goroutines finish in different ways while the GC runs. A finalizer is set so
// the scheduler also collects when idle and clears finished stacks.

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
	for i := 0; i < 48; i++ {
		var batch []*object
		for j := 0; j < 8; j++ {
			batch = append(batch, newObject(-1-j))
		}
		all = append(all, batch)
	}
	runtime.KeepAlive(all)
}

func collect() {
	runtime.GC()
	churn()
}

var (
	mu    sync.Mutex
	count [5]int
)

func bump(kind int) {
	mu.Lock()
	count[kind]++
	mu.Unlock()
}

func returns(id int, wg *sync.WaitGroup) {
	defer wg.Done()
	o := newObject(id)
	runtime.Gosched()
	o.check("returns")
	bump(0)
}

func goexits(id int, wg *sync.WaitGroup) {
	defer wg.Done()
	o := newObject(id)
	defer func() {
		collect()
		o.check("goexit deferred")
		bump(1)
	}()
	func() {
		p := newObject(id + 1)
		defer p.check("goexit inner deferred")
		runtime.Gosched()
		runtime.Goexit()
	}()
	panic("not reached")
}

func recovers(id int, wg *sync.WaitGroup) {
	defer wg.Done()
	o := newObject(id)
	func() {
		defer func() {
			recover()
			collect()
			o.check("recovers")
		}()
		panic("x")
	}()
	bump(2)
}

// collectsLast runs a collection as its final act, while it is finishing.
func collectsLast(id int, wg *sync.WaitGroup) {
	defer wg.Done()
	o := newObject(id)
	runtime.Gosched()
	collect()
	o.check("collects last")
	bump(3)
}

// spawnsAndExits starts a child that outlives it.
func spawnsAndExits(id int, wg *sync.WaitGroup) {
	defer wg.Done()
	o := newObject(id)
	wg.Add(1)
	go func() {
		defer wg.Done()
		p := newObject(id + 1)
		runtime.Gosched()
		collect()
		p.check("child")
		bump(4)
	}()
	o.check("parent")
}

type marker struct{ x int }

func main() {
	// Registering a finalizer enables the idle collection and the clearing of
	// finished stacks.
	m := &marker{1}
	runtime.SetFinalizer(m, func(*marker) {})

	const waves = 25
	var wg sync.WaitGroup
	for w := 0; w < waves; w++ {
		for i := 0; i < 4; i++ {
			id := w*1000 + i*100
			wg.Add(5)
			go returns(id, &wg)
			go goexits(id+10, &wg)
			go recovers(id+20, &wg)
			go collectsLast(id+30, &wg)
			go spawnsAndExits(id+40, &wg)
		}
		if w%5 == 0 {
			// Let the scheduler go idle between waves.
			time.Sleep(2 * time.Millisecond)
		}
		collect()
	}
	wg.Wait()
	collect()
	runtime.KeepAlive(m)
	println("returned:", count[0], "goexit:", count[1], "recovered:", count[2], "collected:", count[3], "children:", count[4])
}
