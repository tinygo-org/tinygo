package main

func testRangeString() {
	for i, c := range "abcü¢€𐍈°x" {
		println(i, c)
	}
}

func testStringToRunes() {
	var s = "abcü¢€𐍈°x"
	for i,c := range []rune(s) {
		println(i, c)
	}
}

func main() {
	testRangeString()
	testStringToRunes()
}
