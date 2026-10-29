package main

import (
	"errors"
	"runtime"
	"sync"
	"time"
)

// Panics are recovered at different depths, re-panicked and replaced, while
// every level holds an object and the deferred functions park and collect.

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

// Baremetal targets report GOARCH arm and have little RAM.
var small = runtime.GOARCH == "arm"

func churn() {
	batches := 48
	if small {
		batches = 6
	}
	var all [][]*object
	for i := 0; i < batches; i++ {
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

func park(i int) {
	switch i % 3 {
	case 0:
		runtime.Gosched()
	case 1:
		time.Sleep(50 * time.Microsecond)
	}
	collect()
}

var errSentinel = errors.New("sentinel")

// down panics at the bottom. Every frame has an object and a deferred check.
//
//go:noinline
func down(depth, id int) {
	o := newObject(id*100 + depth)
	defer func() {
		park(depth)
		o.check("down deferred")
	}()
	if depth == 0 {
		panic(errSentinel)
	}
	down(depth-1, id)
	o.check("down frame")
}

// recoverAt recovers a panic that started depth frames below.
//
//go:noinline
func recoverAt(depth, id int) (err any) {
	o := newObject(id*100 + 90)
	defer func() {
		err = recover()
		park(depth)
		o.check("recoverAt")
	}()
	down(depth, id)
	return nil
}

// repanic recovers and panics again with a new value, several times.
//
//go:noinline
func repanic(level, id int) {
	o := newObject(id*100 + 80 + level)
	defer func() {
		r := recover()
		park(level)
		o.check("repanic")
		panic(r)
	}()
	if level == 0 {
		down(3, id)
		return
	}
	repanic(level-1, id)
}

// replace panics in a deferred function while another panic is running.
//
//go:noinline
func replace(id int) (got any) {
	o := newObject(id*100 + 70)
	defer func() {
		got = recover()
		park(1)
		o.check("replace outer")
	}()
	defer func() {
		p := newObject(id*100 + 71)
		park(2)
		p.check("replace inner")
		panic("second")
	}()
	panic("first")
}

//go:noinline
func deepAfterRecover(depth, id int) int {
	o := newObject(id*100 + 60 + depth)
	if depth == 0 {
		park(id)
		return 0
	}
	n := deepAfterRecover(depth-1, id)
	o.check("after recover frame")
	return n + 1
}

func worker(id int, wg *sync.WaitGroup, results *[4]int, mu *sync.Mutex) {
	defer wg.Done()
	keep := newObject(id*100 + 99)
	recovered, repanics, replaced, loops := 0, 0, 0, 0

	for depth := 1; depth <= 6; depth++ {
		if recoverAt(depth, id) == errSentinel {
			recovered++
		}
		keep.check("keep after recoverAt")
	}

	func() {
		defer func() {
			if recover() == errSentinel {
				repanics++
			}
			park(id)
			keep.check("keep after repanic")
		}()
		repanic(3, id)
	}()

	if replace(id) == "second" {
		replaced++
	}

	// Recovered panics must not leave state behind that confuses later calls.
	if deepAfterRecover(8, id) == 8 {
		loops++
	}
	for i := 0; i < 40; i++ {
		if i%4 == 0 {
			park(i)
		}
		if recoverAt(1+i%4, id) == errSentinel {
			loops++
		}
	}
	keep.check("keep at end")

	mu.Lock()
	results[0] += recovered
	results[1] += repanics
	results[2] += replaced
	results[3] += loops
	mu.Unlock()
}

func main() {
	var wg sync.WaitGroup
	var mu sync.Mutex
	var results [4]int
	for id := 1; id <= 12; id++ {
		wg.Add(1)
		go worker(id, &wg, &results, &mu)
	}
	wg.Wait()
	println("recovered:", results[0], "repanicked:", results[1], "replaced:", results[2], "loops:", results[3])
}
