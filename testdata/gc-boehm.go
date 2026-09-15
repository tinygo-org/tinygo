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

//go:noinline
func makeFalseRoot(setRoot func(uintptr)) {
	object := new(falseRootObject)
	address := uintptr(unsafe.Pointer(object))
	setRoot(address)
	falseRootAddress = address
}

func expectCollected(setRoot func(uintptr)) {
	makeFalseRoot(setRoot)
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
	expectCollected(func(address uintptr) {
		inlineRoot = &inlineFalseRoot{
			value: address,
			live:  new(byte),
		}
	})
	expectCollected(func(address uintptr) {
		externalRoot = &externalFalseRoot{
			value: address,
			live:  new(byte),
		}
	})
	expectCollected(func(address uintptr) {
		repeatedRoots = make([]inlineFalseRoot, 128)
		repeatedRoots[100] = inlineFalseRoot{
			value: address,
			live:  new(byte),
		}
	})
	expectRepeatedPointersLive()
	println("ok")
}
