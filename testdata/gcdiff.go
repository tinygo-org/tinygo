package main

import (
	"runtime"
	"sync"
	"time"
)

// Randomized goroutines park in many ways while holding heap objects. The output
// does not depend on scheduling, so it must match the output with -gc=leaking.

type node struct {
	id      int
	next    *node
	payload [6]int
}

//go:noinline
func newNode(id int, next *node) *node {
	n := &node{id: id, next: next}
	for i := range n.payload {
		n.payload[i] = id*7 + i
	}
	return n
}

func (n *node) verify(what string) int {
	sum := 0
	for ; n != nil; n = n.next {
		for i, v := range n.payload {
			if v != n.id*7+i {
				panic(what + ": corrupted node")
			}
		}
		sum += n.id
	}
	return sum
}

type rng uint32

func (r *rng) next() int {
	x := uint32(*r)
	x ^= x << 13
	x ^= x >> 17
	x ^= x << 5
	*r = rng(x)
	return int(x >> 1)
}

func churn() {
	var all [][]*node
	for i := 0; i < 48; i++ {
		var batch []*node
		for j := 0; j < 8; j++ {
			batch = append(batch, newNode(-1-j, nil))
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
	mu       sync.Mutex
	critical int
	mail     = make(chan *node, 8)
)

// deepPark keeps a node in every frame and parks at the bottom.
//
//go:noinline
func deepPark(depth int, r *rng) int {
	n := newNode(depth+500, nil)
	switch {
	case depth > 0:
		deepPark(depth-1, r)
	case r.next()%2 == 0:
		runtime.Gosched()
		collect()
	default:
		time.Sleep(50 * time.Microsecond)
		collect()
	}
	return n.verify("deep frame")
}

//go:noinline
func panicRecover(depth int, keep *node) {
	defer func() {
		recover()
		keep.verify("panic keep")
	}()
	panicDown(depth, keep)
}

//go:noinline
func panicDown(depth int, keep *node) {
	local := newNode(depth+900, keep)
	if depth == 0 {
		collect()
		panic("diff")
	}
	panicDown(depth-1, keep)
	local.verify("panic frame")
}

func worker(seed int, wg *sync.WaitGroup, total *int, tmu *sync.Mutex) {
	defer wg.Done()
	r := rng(uint32(seed)*2654435761 + 1)
	var list *node
	for i := 0; i < 1+r.next()%5; i++ {
		list = newNode(seed*100+i, list)
	}
	checksum := list.verify("initial")
	criticals := 0

	for step := 0; step < 24; step++ {
		switch r.next() % 9 {
		case 0:
			runtime.Gosched()
		case 1:
			time.Sleep(time.Duration(1+r.next()%3) * 100 * time.Microsecond)
		case 2:
			collect()
		case 3:
			deepPark(r.next()%12, &r)
		case 4:
			panicRecover(1+r.next()%5, list)
		case 5:
			select {
			case m := <-mail:
				m.verify("mail")
				runtime.Gosched()
				collect()
				m.verify("mail after park")
			default:
				select {
				case mail <- newNode(seed*100+50+step, nil):
				default:
				}
			}
		case 6:
			mu.Lock()
			critical++
			criticals++
			runtime.Gosched()
			mu.Unlock()
		case 7:
			select {
			case m := <-mail:
				m.verify("mail select")
			case <-time.After(200 * time.Microsecond):
			}
		case 8:
			churn()
		}
		if list.verify("step") != checksum {
			panic("list changed")
		}
	}

	tmu.Lock()
	*total += checksum + criticals
	tmu.Unlock()
}

func main() {
	const workers = 60
	var wg sync.WaitGroup
	var tmu sync.Mutex
	total := 0
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go worker(i+1, &wg, &total, &tmu)
	}
	wg.Wait()
	println("checksum:", total)
	println("critical sections:", critical)
}
