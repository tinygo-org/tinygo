package main

import "unsafe"

type MySlice [32]byte

type myUint8 uint8

type RecursiveSlice []RecursiveSlice

// Indexing into slice with named type (regression test).
var array = [4]int{
	myUint8(2): 3,
}

func main() {
	l := 5
	foo := []int{1, 2, 4, 5}
	bar := make([]int, l-2, l)
	println("foo is nil?", foo == nil, nil == foo)
	printslice("foo", foo)
	printslice("bar", bar)
	printslice("foo[1:2]", foo[1:2])
	println("sum foo:", sum(foo))

	// creating a slice of uncommon base type
	assert(len(make([]struct{}, makeInt(4))) == 4)

	// creating a slice with uncommon len, cap types
	assert(len(make([]int, makeInt(2), makeInt(3))) == 2)
	assert(len(make([]int, makeInt8(2), makeInt8(3))) == 2)
	assert(len(make([]int, makeInt16(2), makeInt16(3))) == 2)
	assert(len(make([]int, makeInt32(2), makeInt32(3))) == 2)
	assert(len(make([]int, makeInt64(2), makeInt64(3))) == 2)
	assert(len(make([]int, makeUint(2), makeUint(3))) == 2)
	assert(len(make([]int, makeUint8(2), makeUint8(3))) == 2)
	assert(len(make([]int, makeUint16(2), makeUint16(3))) == 2)
	assert(len(make([]int, makeUint32(2), makeUint32(3))) == 2)
	assert(len(make([]int, makeUint64(2), makeUint64(3))) == 2)
	assert(len(make([]int, makeUintptr(2), makeUintptr(3))) == 2)
	assert(len(make([]int, makeMyUint8(2), makeMyUint8(3))) == 2)

	// indexing into a slice with uncommon index types
	assert(foo[int(2)] == 4)
	assert(foo[int8(2)] == 4)
	assert(foo[int16(2)] == 4)
	assert(foo[int32(2)] == 4)
	assert(foo[int64(2)] == 4)
	assert(foo[uint(2)] == 4)
	assert(foo[uint8(2)] == 4)
	assert(foo[uint16(2)] == 4)
	assert(foo[uint32(2)] == 4)
	assert(foo[uint64(2)] == 4)
	assert(foo[uintptr(2)] == 4)

	// slicing with uncommon low, high types
	assert(len(foo[int(1):int(3)]) == 2)
	assert(len(foo[int8(1):int8(3)]) == 2)
	assert(len(foo[int16(1):int16(3)]) == 2)
	assert(len(foo[int32(1):int32(3)]) == 2)
	assert(len(foo[int64(1):int64(3)]) == 2)
	assert(len(foo[uint(1):uint(3)]) == 2)
	assert(len(foo[uint8(1):uint8(3)]) == 2)
	assert(len(foo[uint16(1):uint16(3)]) == 2)
	assert(len(foo[uint32(1):uint32(3)]) == 2)
	assert(len(foo[uint64(1):uint64(3)]) == 2)
	assert(len(foo[uintptr(1):uintptr(3)]) == 2)

	// slicing an array with uncommon low, high types
	arr := [4]int{1, 2, 4, 5}
	assert(len(arr[int(1):int(3)]) == 2)
	assert(len(arr[int8(1):int8(3)]) == 2)
	assert(len(arr[int16(1):int16(3)]) == 2)
	assert(len(arr[int32(1):int32(3)]) == 2)
	assert(len(arr[int64(1):int64(3)]) == 2)
	assert(len(arr[uint(1):uint(3)]) == 2)
	assert(len(arr[uint8(1):uint8(3)]) == 2)
	assert(len(arr[uint16(1):uint16(3)]) == 2)
	assert(len(arr[uint32(1):uint32(3)]) == 2)
	assert(len(arr[uint64(1):uint64(3)]) == 2)
	assert(len(arr[uintptr(1):uintptr(3)]) == 2)

	// slicing with max parameter (added in Go 1.2)
	longfoo := []int{1, 2, 4, 5, 10, 11}
	assert(cap(longfoo[int(1):int(3):int(5)]) == 4)
	assert(cap(longfoo[int8(1):int8(3):int8(5)]) == 4)
	assert(cap(longfoo[int16(1):int16(3):int16(5)]) == 4)
	assert(cap(longfoo[int32(1):int32(3):int32(5)]) == 4)
	assert(cap(longfoo[int64(1):int64(3):int64(5)]) == 4)
	assert(cap(longfoo[uint(1):uint(3):uint(5)]) == 4)
	assert(cap(longfoo[uint8(1):uint8(3):uint8(5)]) == 4)
	assert(cap(longfoo[uint16(1):uint16(3):uint16(5)]) == 4)
	assert(cap(longfoo[uint32(1):uint32(3):uint32(5)]) == 4)
	assert(cap(longfoo[uint64(1):uint64(3):uint64(5)]) == 4)
	assert(cap(longfoo[uintptr(1):uintptr(3):uintptr(5)]) == 4)

	// slicing an array with max parameter (added in Go 1.2)
	assert(cap(arr[int(1):int(2):int(4)]) == 3)
	assert(cap(arr[int8(1):int8(2):int8(4)]) == 3)
	assert(cap(arr[int16(1):int16(2):int16(4)]) == 3)
	assert(cap(arr[int32(1):int32(2):int32(4)]) == 3)
	assert(cap(arr[int64(1):int64(2):int64(4)]) == 3)
	assert(cap(arr[uint(1):uint(2):uint(4)]) == 3)
	assert(cap(arr[uint8(1):uint8(2):uint8(4)]) == 3)
	assert(cap(arr[uint16(1):uint16(2):uint16(4)]) == 3)
	assert(cap(arr[uint32(1):uint32(2):uint32(4)]) == 3)
	assert(cap(arr[uint64(1):uint64(2):uint64(4)]) == 3)
	assert(cap(arr[uintptr(1):uintptr(2):uintptr(4)]) == 3)

	// copy
	println("copy foo -> bar:", copy(bar, foo))
	printslice("bar", bar)

	// append
	var grow []int
	println("slice is nil?", grow == nil, nil == grow)
	printslice("grow", grow)
	grow = append(grow, 42)
	printslice("grow", grow)
	grow = append(grow, -1, -2)
	printslice("grow", grow)
	grow = append(grow, foo...)
	printslice("grow", grow)
	grow = append(grow)
	printslice("grow", grow)
	grow = append(grow, grow...)
	printslice("grow", grow)

	// append string to []bytes
	bytes := append([]byte{1, 2, 3}, "foo"...)
	print("bytes: len=", len(bytes), " cap=", cap(bytes), " data:")
	for _, n := range bytes {
		print(" ", n)
	}
	println()

	// Test conversion from array to slice.
	slice1 := []int{1, 2, 3, 4}
	arr1 := (*[4]int)(slice1)
	arr1[1] = -2
	arr1[2] = 20
	println("slice to array pointer:", arr1[0], arr1[1], arr1[2], arr1[3])

	// Test unsafe.Add.
	arr2 := [...]int{1, 2, 3, 4}
	*(*int)(unsafe.Add(unsafe.Pointer(&arr2[0]), unsafe.Sizeof(int(1))*1)) = 5
	*addInt(&arr2[0], 2) = 8
	println("unsafe.Add array:", arr2[0], arr2[1], arr2[2], arr2[3])

	// Test unsafe.Slice.
	arr3 := [...]int{1, 2, 3, 4}
	slice3 := unsafe.Slice(&arr3[1], 3)
	slice3[0] = 9
	slice3[1] = 15
	println("unsafe.Slice array:", len(slice3), cap(slice3), slice3[0], slice3[1], slice3[2])

	// Verify the fix in https://github.com/tinygo-org/tinygo/pull/119
	var unnamed [32]byte
	var named MySlice
	assert(len(unnamed[:]) == 32)
	assert(len(named[:]) == 32)
	for _, c := range named {
		assert(c == 0)
	}

	// Test recursive slices.
	rs := []RecursiveSlice(nil)
	println("len:", len(rs))

	testRangeLoops()
	testSliceBounds()
}

// Range loops evaluate the slice once and keep their own copy of the index.
func testRangeLoops() {
	var nilSlice []int
	empty := make([]int, 0, 8)
	data := []int{10, 20, 30, 40, 50}

	count := func(s []int) (n int) {
		for range s {
			n++
		}
		return
	}
	println("range count:", count(nilSlice), count(empty), count(data), count(data[:0]), count(data[1:]), count(data[:3:3]))

	// Changing the slice or the index inside the loop does not change the
	// number of iterations.
	n := 0
	for i := range data {
		data = append(data, i)
		i += 100
		n++
	}
	println("range append:", n, len(data))

	s := []int{1, 2, 3, 4}
	sum := 0
	for i, v := range s {
		s = nil
		sum += i*10 + v
	}
	println("range reassign:", sum, len(s))

	s = []int{1, 2, 3, 4, 5, 6}
	sum = 0
	for i := range s {
		if i == 1 {
			continue
		}
		if i == 4 {
			break
		}
		sum += s[i]
	}
	println("range continue break:", sum)

	// A range loop over a slice that is changed by the loop body sees the
	// original length but the new elements.
	s = []int{1, 2, 3}
	sum = 0
	for i := range s {
		if i == 0 {
			s[1] = 20
			s[2] = 30
		}
		sum += s[i]
	}
	println("range write:", sum)

	var nested [][]int
	for i := 0; i < 4; i++ {
		nested = append(nested, make([]int, i))
	}
	sum = 0
	for _, row := range nested[1:] {
		for j := range row {
			sum += j + 1
		}
	}
	println("range nested:", sum)
}

var sliceSink []byte

func slicePanics(name string, f func()) {
	defer func() {
		println(name, recover() != nil)
	}()
	f()
}

// Test that slicing a slice still panics for every out of range case.
func testSliceBounds() {
	b := make([]byte, 10, 20)
	low, high, max := makeInt(5), makeInt(15), makeInt(20)
	slicePanics("b[len:]", func() { sliceSink = b[makeInt(10):] })
	slicePanics("b[len+1:]", func() { sliceSink = b[makeInt(11):] })
	slicePanics("b[:cap]", func() { sliceSink = b[:makeInt(20)] })
	slicePanics("b[:cap+1]", func() { sliceSink = b[:makeInt(21)] })
	slicePanics("b[low:high]", func() { sliceSink = b[low:high] })
	slicePanics("b[high:low]", func() { sliceSink = b[high:low] })
	slicePanics("b[low:high:max]", func() { sliceSink = b[low:high:max] })
	slicePanics("b[low:high:cap+1]", func() { sliceSink = b[low:high:makeInt(21)] })
	slicePanics("b[low:max:high]", func() { sliceSink = b[low:max:high] })
	slicePanics("b[-1:]", func() { sliceSink = b[makeInt(-1):] })
	slicePanics("b[:-1]", func() { sliceSink = b[:makeInt(-1)] })
}

func printslice(name string, s []int) {
	print(name, ": len=", len(s), " cap=", cap(s), " data:")
	for _, n := range s {
		print(" ", n)
	}
	println()
}

func sum(l []int) int {
	sum := 0
	for _, n := range l {
		sum += n
	}
	return sum
}

func assert(ok bool) {
	if !ok {
		panic("assert failed")
	}
}

// Helper functions used to hide const values from the compiler during IR
// construction.

func makeInt(x int) int             { return x }
func makeInt8(x int8) int8          { return x }
func makeInt16(x int16) int16       { return x }
func makeInt32(x int32) int32       { return x }
func makeInt64(x int64) int64       { return x }
func makeUint(x uint) uint          { return x }
func makeUint8(x uint8) uint8       { return x }
func makeUint16(x uint16) uint16    { return x }
func makeUint32(x uint32) uint32    { return x }
func makeUint64(x uint64) uint64    { return x }
func makeUintptr(x uintptr) uintptr { return x }
func makeMyUint8(x myUint8) myUint8 { return x }

func addInt(ptr *int, index uintptr) *int {
	return (*int)(unsafe.Add(unsafe.Pointer(ptr), unsafe.Sizeof(int(1))*index))
}
