package main

import "slices"

var names []string
var ids []int
var bytes []byte

//go:noinline
func count(name string, id int, b byte) int {
	n := 0
	if slices.Contains(names, name) {
		n++
	}
	if slices.Contains(ids, id) {
		n++
	}
	if slices.Contains(bytes, b) {
		n++
	}
	return n
}

func main() {
	for i := 0; i < 8; i++ {
		names = append(names, string(rune('a'+i)))
		ids = append(ids, i*i)
		bytes = append(bytes, byte(i))
	}
	println(count("c", 9, 3), count("z", 10, 30))
}
