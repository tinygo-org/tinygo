package main

import "unicode/utf8"

var m = map[string]int{"abc": 1, "axc": 2}

//go:noinline
func lookup(b []byte) int {
	return m[string(b)]
}

//go:noinline
func count(b []byte) int {
	return utf8.RuneCountInString(string(b))
}

func main() {
	// Changing the slice during the loop must not change the iteration.
	b := []byte("a\xc3\xa9")
	for i := range string(b) {
		println(i)
		b[1] = 'x'
	}

	// The conversion is a snapshot, so later writes are not visible.
	s := []byte("abc")
	println(lookup(s))
	s[1] = 'x'
	println(lookup(s))
	println(lookup([]byte("none")))

	r := []byte("\xe6\x97\xa5\xe6\x9c\xac")
	println(count(r))
	r[0] = 'a'
	println(count(r))
}
