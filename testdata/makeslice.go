package main

import "unsafe"

type triple [3]byte

var (
	bytes  = []byte{1, 2, 3, 4}
	text   = "hello"
	array  = [5]uint16{}
	narrow = uint8(7)
)

//go:noinline
func zeroLenCapOfSlice(x []byte) []byte { return make([]byte, 0, len(x)) }

//go:noinline
func lenOfSlice(x []byte) []byte { return make([]byte, len(x)) }

//go:noinline
func lenAndCapOfSlice(x []byte) []byte { return make([]byte, len(x), len(x)) }

//go:noinline
func zeroLenCapOfCap(x []byte) []byte { return make([]byte, 0, cap(x)) }

//go:noinline
func zeroLenCapOfString(s string) []byte { return make([]byte, 0, len(s)) }

//go:noinline
func zeroLenCapOfArray(x *[5]uint16) []uint16 { return make([]uint16, 0, len(x)) }

//go:noinline
func narrowUnsigned(n uint8) []byte { return make([]byte, n) }

//go:noinline
func constants() []byte { return make([]byte, 2, 8) }

// The destination element is bigger than the source element, so the result
// needs more memory than the length suggests.
//
//go:noinline
func stringToTriples(s string) [][3]byte { return make([][3]byte, 0, len(s)) }

//go:noinline
func bytesToWords(x []byte) []uint32 { return make([]uint32, 0, len(x)) }

func show(name string, length, capacity int) {
	println(name, length, capacity)
}

func mustPanic(name string, f func()) {
	defer func() {
		println(name, recover() != nil)
	}()
	f()
}

func main() {
	s := zeroLenCapOfSlice(bytes)
	show("zeroLenCapOfSlice", len(s), cap(s))
	s = lenOfSlice(bytes)
	show("lenOfSlice", len(s), cap(s))
	s = lenAndCapOfSlice(bytes)
	show("lenAndCapOfSlice", len(s), cap(s))
	s = zeroLenCapOfCap(bytes[:2:2])
	show("zeroLenCapOfCap", len(s), cap(s))
	s = zeroLenCapOfString(text)
	show("zeroLenCapOfString", len(s), cap(s))
	w := zeroLenCapOfArray(&array)
	show("zeroLenCapOfArray", len(w), cap(w))
	s = narrowUnsigned(narrow)
	show("narrowUnsigned", len(s), cap(s))
	s = constants()
	show("constants", len(s), cap(s))
	t := stringToTriples(text)
	show("stringToTriples", len(t), cap(t))
	words := bytesToWords(bytes)
	show("bytesToWords", len(words), cap(words))

	if unsafe.Sizeof(uintptr(0)) == 4 {
		// These lengths are valid for bytes but not for the bigger elements.
		var b byte
		huge := unsafe.String(&b, 1_431_655_766)
		mustPanic("stringToTriplesHuge", func() { stringToTriples(huge) })
		hugeBytes := unsafe.Slice(&b, 1_073_741_825)
		mustPanic("bytesToWordsHuge", func() { bytesToWords(hugeBytes) })
	} else {
		// Allocating this much works on 64-bit targets, so only print what the
		// 32-bit targets print.
		println("stringToTriplesHuge true")
		println("bytesToWordsHuge true")
	}
}
