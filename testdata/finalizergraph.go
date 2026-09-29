package main

import "runtime"

type graphNode struct {
	next *graphNode
	tag  uint32
	data [128]byte
}

var (
	graphFinalized int
	cycleFinalized int
	graphSink      int
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

//go:noinline
func scrubGraphStack(depth int) int {
	if depth == 0 {
		return graphSink
	}
	var buf [64]int
	for i := range buf {
		buf[i] = depth + i
	}
	return scrubGraphStack(depth-1) + buf[0]
}

func main() {
	registerGraphFinalizer()
	registerCycleFinalizers()
	for i := 0; i < 100 && (graphFinalized != 1 || cycleFinalized != 2); i++ {
		graphSink += scrubGraphStack(40)
		runtime.GC()
		runtime.Gosched()
	}
	if graphFinalized != 1 || cycleFinalized != 2 {
		panic("finalizers did not run for the graph and cycle")
	}
	println("ok")
}
