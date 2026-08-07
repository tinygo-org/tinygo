package main

import (
	"reflect"
	"runtime"
)

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
	testGlobalMapRoots()
	testGlobalChannelRoots()
	testReflectRoots()
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

type globalMapObject struct {
	marker int
	data   [64]byte
}

var globalMap = make(map[int]*globalMapObject)
var globalChannel chan *globalMapObject
var globalPointerSlice []*globalMapObject
var globalGCClobber any

type globalMapLargeKey struct {
	object *globalMapObject
	data   [129]byte
}

type globalMapLargeValue struct {
	object *globalMapObject
	data   [129]byte
}

var globalLargeKeyMap = make(map[globalMapLargeKey]int)
var globalLargeValueMap = make(map[int]globalMapLargeValue)

//go:noinline
func populateGlobalMaps() {
	for i := 0; i < 32; i++ {
		globalMap[i] = &globalMapObject{marker: 100 + i}
		globalPointerSlice = append(globalPointerSlice, &globalMapObject{marker: 800 + i})
	}
	globalLargeKeyMap[globalMapLargeKey{
		object: &globalMapObject{marker: 200},
	}] = 1
	globalLargeValueMap[0] = globalMapLargeValue{
		object: &globalMapObject{marker: 300},
	}
}

func testGlobalMapRoots() {
	populateGlobalMaps()

	runtime.GC()
	for i := 0; i < 100; i++ {
		globalGCClobber = new(globalMapObject)
	}
	runtime.GC()

	for i := 0; i < 32; i++ {
		if globalMap[i].marker != 100+i {
			panic("global map value was collected")
		}
	}
	for key := range globalLargeKeyMap {
		if key.object.marker != 200 {
			panic("indirect global map key was collected")
		}
	}
	if globalLargeValueMap[0].object.marker != 300 {
		panic("indirect global map value was collected")
	}
	for i, object := range globalPointerSlice {
		if object.marker != 800+i {
			panic("global slice value was collected")
		}
	}
}

//go:noinline
func populateGlobalChannel() {
	globalChannel = make(chan *globalMapObject, 4)
	globalChannel <- &globalMapObject{marker: 400}
}

func testGlobalChannelRoots() {
	populateGlobalChannel()

	runtime.GC()
	for i := 0; i < 100; i++ {
		globalGCClobber = new(globalMapObject)
	}
	runtime.GC()

	if (<-globalChannel).marker != 400 {
		panic("global channel value was collected")
	}
}

type reflectRootObject struct {
	marker int
	child  *reflectRootObject
	data   [128]byte
}

type reflectMapKey struct {
	object *reflectRootObject
	data   [129]byte
}

type reflectMapValue struct {
	object *reflectRootObject
	data   [129]byte
}

type reflectRootMap map[reflectMapKey]reflectMapValue

var globalReflectObject *reflectRootObject
var globalReflectMap reflectRootMap

//go:noinline
func populateReflectRoots() {
	value := reflect.New(reflect.TypeOf(reflectRootObject{}))
	globalReflectObject = value.Interface().(*reflectRootObject)
	globalReflectObject.child = &reflectRootObject{marker: 500}

	mapValue := reflect.MakeMapWithSize(reflect.TypeOf(globalReflectMap), 1)
	mapValue.SetMapIndex(
		reflect.ValueOf(reflectMapKey{object: &reflectRootObject{marker: 600}}),
		reflect.ValueOf(reflectMapValue{object: &reflectRootObject{marker: 700}}),
	)
	globalReflectMap = mapValue.Interface().(reflectRootMap)
}

func testReflectRoots() {
	populateReflectRoots()

	runtime.GC()
	for i := 0; i < 100; i++ {
		globalGCClobber = new(reflectRootObject)
	}
	runtime.GC()

	if globalReflectObject.child.marker != 500 {
		panic("reflected object field was collected")
	}
	for key, value := range globalReflectMap {
		if key.object.marker != 600 {
			panic("reflected map key was collected")
		}
		if value.object.marker != 700 {
			panic("reflected map value was collected")
		}
	}
}
