package main

import "runtime"

var xorshift32State uint32 = 1

func xorshift32(x uint32) uint32 {
	// Algorithm "xor" from p. 4 of Marsaglia, "Xorshift RNGs"
	x ^= x << 13
	x ^= x >> 17
	x ^= x << 5
	return x
}

func randuint32() uint32 {
	xorshift32State = xorshift32(xorshift32State)
	return xorshift32State
}

func main() {
	testNonPointerHeap()
	testKeepAlive()
}

var scalarSlices [4][]byte
var randSeeds [4]uint32

func testNonPointerHeap() {
	maxSliceSize := uint32(1024)
	if ^uintptr(0) <= 0xffff {
		// 16-bit and lower devices, such as AVR.
		// Heap size is a real issue there, while it is still useful to run
		// these tests. Therefore, lower the max slice size.
		maxSliceSize = 64
	}
	// Allocate roughly 0.5MB of memory.
	for i := 0; i < 1000; i++ {
		// Pick a random index that the optimizer can't predict.
		index := randuint32() % 4

		// Check whether the contents of the previous allocation was correct.
		rand := randSeeds[index]
		for _, b := range scalarSlices[index] {
			rand = xorshift32(rand)
			if b != byte(rand) {
				panic("memory was overwritten!")
			}
		}

		// Allocate a randomly-sized slice, randomly sliced to be smaller.
		sliceLen := randuint32() % maxSliceSize
		slice := make([]byte, sliceLen)
		cutLen := randuint32() % maxSliceSize
		if cutLen < sliceLen {
			slice = slice[cutLen:]
		}
		scalarSlices[index] = slice

		// Fill the slice with a pattern that looks random but is easily
		// calculated and verified.
		rand = randuint32() + 1
		randSeeds[index] = rand
		for i := 0; i < len(slice); i++ {
			rand = xorshift32(rand)
			slice[i] = byte(rand)
		}
	}
	println("ok")
}

func testKeepAlive() {
	// There isn't much we can test, but at least we can test that
	// runtime.KeepAlive compiles correctly.
	var x int
	runtime.KeepAlive(&x)
}
