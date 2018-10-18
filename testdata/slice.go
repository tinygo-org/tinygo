package main

func main() {
	l := 5
	foo := []int{1, 2, 4, 5}
	bar := make([]int, l-2, l)
	printslice("foo", foo)
	printslice("bar", bar)
	printslice("foo[1:2]", foo[1:2])
	println("sum foo:", sum(foo))
	println("copy foo -> bar:", copy(bar, foo))
	printslice("bar", bar)
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
