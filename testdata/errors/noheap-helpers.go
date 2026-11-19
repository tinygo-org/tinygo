package main

var (
	bs    []byte
	ints  []int
	str   string
	runes []rune
	m     map[int]int
	ch    chan int
	count int
)

//go:noheap
func main() {
	// These operations call a runtime function that can allocate.
	ints = append(ints, 1)
	if len(ints) < cap(ints) {
		ints = append(ints, 2)
	}
	str = str + "x"
	bs = []byte(str)
	str = string(bs)
	runes = []rune(str)
	str = string(runes)
	str = string(rune(count))
	m = make(map[int]int)
	m[1] = 2
	ch = make(chan int, 1)
	go func() { count++ }()

	// These operations do not allocate.
	println(string(bs) == "hello")
	println(m[1])
	println(copy(bs, "zz"))
	for _, r := range str {
		count += int(r)
	}
}

// ERROR: # command-line-arguments
// ERROR: noheap-helpers.go:16:15: append may allocate on the heap in //go:noheap function
// ERROR: noheap-helpers.go:18:16: append may allocate on the heap in //go:noheap function
// ERROR: noheap-helpers.go:20:12: string concatenation may allocate on the heap in //go:noheap function
// ERROR: noheap-helpers.go:21:13: conversion from string to []byte may allocate on the heap in //go:noheap function
// ERROR: noheap-helpers.go:22:14: conversion from []byte to string may allocate on the heap in //go:noheap function
// ERROR: noheap-helpers.go:23:16: conversion from string to []rune may allocate on the heap in //go:noheap function
// ERROR: noheap-helpers.go:24:14: conversion from []rune to string may allocate on the heap in //go:noheap function
// ERROR: noheap-helpers.go:25:14: conversion from rune to string may allocate on the heap in //go:noheap function
// ERROR: noheap-helpers.go:26:10: map creation may allocate on the heap in //go:noheap function
// ERROR: noheap-helpers.go:27:3: map assignment may allocate on the heap in //go:noheap function
// ERROR: noheap-helpers.go:28:11: channel creation may allocate on the heap in //go:noheap function
// ERROR: noheap-helpers.go:29:2: go statement may allocate on the heap in //go:noheap function
