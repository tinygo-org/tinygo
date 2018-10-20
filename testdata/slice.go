package main

func main() {
	l := 5
	foo := []int{1, 2, 4, 5}
	bar := make([]int, l-2, l)
	println("foo is nil?", foo == nil, nil == foo)
	printslice("foo", foo)
	printslice("bar", bar)
	printslice("foo[1:2]", foo[1:2])
	println("sum foo:", sum(foo))

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
