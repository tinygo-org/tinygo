package main

import "runtime"

type referentPayload struct {
	data [256]byte
}

type referentOwner struct {
	payload *referentPayload
}

type referentCycleNode struct {
	peer *referentCycleNode
	data [64]byte
}

const (
	referentOwners = 64
	referentCycles = 32
	referentFill   = 0xAA
)

var (
	referentRan       int
	referentCorrupted int
	referentCycleRan  int
	referentKeep      [][]byte
	sink              int
)

// scrubStack removes stale pointers from the helper frame so collection is deterministic.
//
//go:noinline
func scrubStack(depth int) int {
	if depth <= 0 {
		return sink
	}
	var buf [64]int
	for i := range buf {
		buf[i] = depth + i
	}
	sink += buf[depth&63]
	return scrubStack(depth-1) + buf[0]
}

// Each payload is reachable only through its owner.
//
//go:noinline
func registerReferentOwners() {
	for i := 0; i < referentOwners; i++ {
		owner := &referentOwner{payload: &referentPayload{}}
		for j := range owner.payload.data {
			owner.payload.data[j] = referentFill
		}
		runtime.SetFinalizer(owner, func(owner *referentOwner) {
			referentRan++
			// Allocate while the payload is read, like a finalizer that does work.
			allocatePayloadSized(64)
			for _, b := range owner.payload.data {
				if b != referentFill {
					referentCorrupted++
					return
				}
			}
		})
	}
}

// Allocate blocks of the payload size so freed payloads are used again.
//
//go:noinline
func allocatePayloadSized(count int) {
	for i := 0; i < count; i++ {
		block := make([]byte, 256)
		for j := range block {
			block[j] = 0x55
		}
		referentKeep = append(referentKeep, block)
	}
}

// Both objects of each cycle have a finalizer. No order is required.
//
//go:noinline
func registerReferentCycles() {
	for i := 0; i < referentCycles; i++ {
		a, b := &referentCycleNode{}, &referentCycleNode{}
		a.peer, b.peer = b, a
		runtime.SetFinalizer(a, func(*referentCycleNode) { referentCycleRan++ })
		runtime.SetFinalizer(b, func(*referentCycleNode) { referentCycleRan++ })
	}
}

func main() {
	registerReferentOwners()
	registerReferentCycles()
	sink += scrubStack(40)
	// The runner does not get a turn before the blocks are used again.
	runtime.GC()
	allocatePayloadSized(4000)
	for i := 0; i < 100 && (referentRan == 0 || referentCycleRan < 2*referentCycles); i++ {
		sink += scrubStack(40)
		runtime.GC()
		runtime.Gosched()
	}
	if referentRan == 0 {
		panic("no finalizer ran")
	}
	if referentCorrupted != 0 {
		println("finalizers ran", referentRan, "and saw a freed payload", referentCorrupted)
		panic("finalizer saw a freed payload")
	}
	// A stale stack pointer can keep one cycle alive.
	if referentCycleRan < 2*referentCycles-2 {
		println("cycle finalizers ran", referentCycleRan, "of", 2*referentCycles)
		panic("cycle finalizers did not run")
	}
	println("ok")
}
