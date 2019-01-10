package main

type MySlice [32]byte

func returnUint64() uint64 {
	return uint64(5)
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

	// creating a slice with uncommon len, cap types
	assert(len(make([]int, int(2), int(3))) == 2)
	assert(len(make([]int, int8(2), int8(3))) == 2)
	assert(len(make([]int, int16(2), int16(3))) == 2)
	assert(len(make([]int, int32(2), int32(3))) == 2)
	assert(len(make([]int, int64(2), int64(3))) == 2)
	assert(len(make([]int, uint(2), uint(3))) == 2)
	assert(len(make([]int, uint8(2), uint8(3))) == 2)
	assert(len(make([]int, uint16(2), uint16(3))) == 2)
	assert(len(make([]int, uint32(2), uint32(3))) == 2)
	assert(len(make([]int, uint64(2), uint64(3))) == 2)
	assert(len(make([]int, uintptr(2), uintptr(3))) == 2)
	assert(len(make([]int, returnUint64(), returnUint64())) == 5)

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

	// Verify the fix in https://github.com/aykevl/tinygo/pull/119
	var unnamed [32]byte
	var named MySlice
	assert(len(unnamed[:]) == 32)
	assert(len(named[:]) == 32)
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
