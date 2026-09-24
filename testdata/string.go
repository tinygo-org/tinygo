package main

func testRangeString() {
	for i, c := range "abcü¢€𐍈°x" {
		println(i, c)
	}
}

func testStringToRunes() {
	var s = "abcü¢€𐍈°x"
	for i, c := range []rune(s) {
		println(i, c)
	}
}

func testRunesToString(r []rune) {
	println("string from runes:", string(r))
}

func testByteSliceStringCompareNil() {
	var nilSlice []byte
	empty := []byte{}
	full := []byte("foo")
	println(string(nilSlice) == string(empty))
	println(string(nilSlice) == string(full))
	println(string(empty) == string(nilSlice))
	println(string(nilSlice) != string(empty))
}

type myString string

func main() {
	testRangeString()
	testStringToRunes()
	testRunesToString([]rune{97, 98, 99, 252, 162, 8364, 66376, 176, 120})
	testByteSliceStringCompareNil()
	var _ = len([]byte(myString("foobar"))) // issue 1246
}
