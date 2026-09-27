package main

import (
	"runtime"
	"unsafe"
)

type falseRootObject struct {
	data [64]byte
}

type inlineFalseRoot struct {
	value uintptr
	live  *byte
}

type externalFalseRoot struct {
	value   uintptr
	padding [64]uintptr
	live    *byte
}

var inlineRoot *inlineFalseRoot
var externalRoot *externalFalseRoot
var repeatedRoots []inlineFalseRoot
var falseRootAddress uintptr
var frameSum uintptr

//go:noinline
func fillFrame(frame []uintptr) {
	for i := range frame {
		frame[i] = uintptr(i)
		frameSum += frame[i]
	}
}

// newFalseRootObject allocates depth frames down, leaving stale copies of
// the pointer below the stack pointer of the caller.
//
//go:noinline
func newFalseRootObject(depth int) *falseRootObject {
	var frame [4]uintptr
	fillFrame(frame[:])
	if depth == 0 {
		return new(falseRootObject)
	}
	object := newFalseRootObject(depth - 1)
	frameSum += frame[0]
	return object
}

//go:noinline
func makeFalseRoot(depth int, setRoot func(uintptr)) {
	object := newFalseRootObject(depth)
	address := uintptr(unsafe.Pointer(object))
	setRoot(address)
	falseRootAddress = address
}

func expectCollected(depth int, setRoot func(uintptr)) {
	done := make(chan struct{})
	go func() {
		makeFalseRoot(depth, setRoot)
		close(done)
	}()
	<-done
	runtime.GC()
	for i := 0; i < 100000; i++ {
		object := new(falseRootObject)
		if uintptr(unsafe.Pointer(object)) == falseRootAddress {
			return
		}
	}
	panic("non-pointer field retained allocation")
}

func expectRepeatedPointersLive() {
	repeatedRoots = make([]inlineFalseRoot, 128)
	for i := range repeatedRoots {
		value := new(byte)
		*value = byte(i)
		repeatedRoots[i].live = value
	}
	runtime.GC()
	for i := range repeatedRoots {
		if *repeatedRoots[i].live != byte(i) {
			panic("repeated pointer field was not retained")
		}
	}
}

func main() {
	expectCollected(0, func(address uintptr) {
		inlineRoot = &inlineFalseRoot{
			value: address,
			live:  new(byte),
		}
	})
	expectCollected(0, func(address uintptr) {
		externalRoot = &externalFalseRoot{
			value: address,
			live:  new(byte),
		}
	})
	expectCollected(0, func(address uintptr) {
		repeatedRoots = make([]inlineFalseRoot, 128)
		repeatedRoots[100] = inlineFalseRoot{
			value: address,
			live:  new(byte),
		}
	})
	expectCollected(8, func(address uintptr) {
		inlineRoot = &inlineFalseRoot{
			value: address,
			live:  new(byte),
		}
	})
	expectRepeatedPointersLive()
	println("ok")
}
