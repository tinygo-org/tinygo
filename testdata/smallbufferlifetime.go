package main

// The buffer made for v is read after string(v) is copied, so its lifetime
// must not end in the copy.

var sinkSum int

//go:noinline
func count(s string) int {
	n := 0
	for i := 0; i < len(s); i++ {
		n += int(s[i])
	}
	return n
}

//go:noinline
func sum(b []byte) int {
	n := 0
	for _, c := range b {
		n += int(c)
	}
	return n
}

//go:noinline
func noop() {}

func stringThenBytes(n int) {
	// A deferred call makes every call end its block, so the uses of s and v
	// are in blocks after the one that defines them.
	defer noop()
	v := make([]byte, n)
	s := string(v)
	sinkSum += count(s)
	sinkSum += sum(v)
}

func main() {
	// Two sizes, so the size is not a constant.
	stringThenBytes(29)
	stringThenBytes(40)
}
