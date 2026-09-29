package main

import (
	"os"
	"runtime"
)

type largeFinalizerObject struct {
	data [128]byte
}

var largeFinalizerRan bool
var largeFinalizerSink int

//go:noinline
func scrubLargeFinalizerStack(depth int) int {
	if depth == 0 {
		return largeFinalizerSink
	}
	var buf [64]int
	for i := range buf {
		buf[i] = depth + i
	}
	return scrubLargeFinalizerStack(depth-1) + buf[0]
}

//go:noinline
func registerLargeFinalizer() {
	p := new(largeFinalizerObject)
	payload := new(largeFinalizerObject)
	for i := range payload.data {
		payload.data[i] = byte(i)
	}
	runtime.SetFinalizer(p, func(*largeFinalizerObject) {
		for i, value := range payload.data {
			if value != byte(i) {
				panic("finalizer closure data was collected")
			}
		}
		largeFinalizerRan = true
	})
}

func main() {
	if len(os.Args) > 1 && os.Args[1] == "graph" {
		testFinalizerGraph()
		println("ok")
		return
	}
	registerLargeFinalizer()
	for i := 0; i < 100 && !largeFinalizerRan; i++ {
		largeFinalizerSink += scrubLargeFinalizerStack(40)
		runtime.GC()
		runtime.Gosched()
	}
	if !largeFinalizerRan {
		panic("large object finalizer did not run")
	}
	println("ok")
}

type graphNode struct {
	next *graphNode
	tag  uint32
	data [128]byte
}

var (
	graphFinalized int
	cycleFinalized int
	graphKeepAlive []*graphNode
)

//go:noinline
func registerGraphFinalizer() {
	child := &graphNode{tag: 0x12345678}
	parent := &graphNode{next: child}
	runtime.SetFinalizer(parent, func(p *graphNode) {
		for i := 0; i < 8192; i++ {
			graphKeepAlive = append(graphKeepAlive, &graphNode{tag: uint32(i)})
		}
		if p.next.tag != 0x12345678 {
			panic("finalizer lost a referenced object")
		}
		graphFinalized++
	})
}

//go:noinline
func registerCycleFinalizers() {
	first := new(graphNode)
	second := new(graphNode)
	first.next = second
	second.next = first
	runtime.SetFinalizer(first, func(*graphNode) { cycleFinalized++ })
	runtime.SetFinalizer(second, func(*graphNode) { cycleFinalized++ })
}

func testFinalizerGraph() {
	registerGraphFinalizer()
	registerCycleFinalizers()
	for i := 0; i < 100 && (graphFinalized != 1 || cycleFinalized != 2); i++ {
		largeFinalizerSink += scrubLargeFinalizerStack(40)
		runtime.GC()
		runtime.Gosched()
	}
	if graphFinalized != 1 || cycleFinalized != 2 {
		panic("finalizers did not run for the graph and cycle")
	}
}
