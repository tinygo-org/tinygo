package main

func main() {
	// test basic printing
	println("hello world!")
	println(42)
	println(100000000)

	// check that this one doesn't print an extra space between args
	print("a", "b", "c")
	println()
	// ..but this one does
	println("a", "b", "c")

	// print integers
	println(uint8(123))
	println(int8(123))
	println(int8(-123))
	println(uint16(12345))
	println(int16(12345))
	println(int16(-12345))
	println(uint32(12345678))
	println(int32(12345678))
	println(int32(-12345678))
	println(uint64(123456789012))
	println(int64(123456789012))
	println(int64(-123456789012))

	// print float64
	println(3.14)

	// print float32
	println(float32(3.14))

	// print complex128
	println(5 + 1.2345i)

	// print interface
	println(interface{}(nil))

	// print map
	println(map[string]int{"three": 3, "five": 5})

	// TODO: print pointer

	// print bool
	println(true, false)

	// print slice
	println([]byte(nil))
	println([]int(nil))
}
