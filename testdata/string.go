package main

import "bytes"

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

func testByteSliceStringCompareEmpty() {
	var zeroCap []byte
	zeroLen := make([]byte, 0, 1)
	backing := []byte{'x'}
	zeroLenWithBacking := backing[:0]
	uint8Empty := []uint8{}

	println("empty:", string(zeroCap) == string(zeroLen),
		string(zeroLen) == string(zeroLenWithBacking),
		string(zeroLenWithBacking) == string(uint8Empty))
}

type myString string

type myByte byte
type myBytes []myByte
type myRune rune

//go:noinline
func compareByteStrings(a, b []byte) (bool, bool) {
	return string(a) == string(b), string(a) != string(b)
}

//go:noinline
func compareLocalByteStrings(a, b []byte) bool {
	s := string(a)
	t := string(b)
	return s == t
}

//go:noinline
func escapeByteString(a, b []byte) (bool, string) {
	s := string(a)
	t := string(b)
	return s == t, s
}

//go:noinline
func storeByteString(a, b []byte, dst *string) bool {
	s := string(a)
	t := string(b)
	equal := s == t
	*dst = s
	return equal
}

//go:noinline
func boxByteString(a, b []byte) (bool, any) {
	s := string(a)
	t := string(b)
	return s == t, any(s)
}

//go:noinline
func mutateByteString(a []byte) (bool, string) {
	s := string(a)
	a[0] = 'z'
	return s == string(a), s
}

//go:noinline
func compareByteStringsAfterStore(a, b []byte) bool {
	s := string(a)
	b[0] = 'z'
	return s == string(b)
}

//go:noinline
func mutateByteSlice(a []byte) {
	a[0] = 'z'
}

//go:noinline
func compareByteStringsAfterCall(a, b []byte) bool {
	s := string(a)
	mutateByteSlice(b)
	return s == string(b)
}

//go:noinline
func compareByteStringsAcrossBlock(a, b []byte, mutate bool) bool {
	s := string(a)
	if mutate {
		b[0] = 'z'
	}
	return s == string(b)
}

//go:noinline
func compareNamedByteStrings(a, b myBytes) (bool, bool) {
	return string(a) == string(b), myString(a) != myString(b)
}

//go:noinline
func convertNamedBytes(a myBytes) (string, myString) {
	return string(a), myString(a)
}

//go:noinline
func convertNamedRunes(a []myRune) string {
	return string(a)
}

//go:noinline
func compareByteStringFallbacks(a, b []byte, s string) {
	println("fallbacks:", string(a) == s, string(a) == "abc",
		string(a[:2]) == string(b[:2]), string(a) < s)
	var x, y any = string(a), string(b)
	a[0] = 'z'
	println("boxed:", x == y, x.(string), y.(string))
}

//go:noinline
func findBytes(a, b []byte) int {
	return bytes.Index(a, b)
}

func testByteSliceStringComparisons() {
	for _, tc := range []struct{ a, b []byte }{
		{nil, nil},
		{nil, []byte{}},
		{[]byte{}, nil},
		{nil, []byte{'a'}},
		{[]byte{'a'}, []byte{'a'}},
		{[]byte{'a'}, []byte{'a', 'a'}},
		{[]byte{'a', 'b'}, []byte{'a', 'c'}},
		{[]byte{0, 0xff}, []byte{0, 0xff}},
	} {
		equal, unequal := compareByteStrings(tc.a, tc.b)
		println("compare:", equal, unequal, compareLocalByteStrings(tc.a, tc.b))
	}

	a := []byte("abc")
	b := []byte("abc")
	equal, snapshot := escapeByteString(a, b)
	a[0] = 'z'
	println("escape:", equal, snapshot, string(a))
	a[0] = 'a'
	equal = storeByteString(a, b, &snapshot)
	a[0] = 'z'
	println("store:", equal, snapshot, string(a))
	a[0] = 'a'
	equal, boxed := boxByteString(a, b)
	a[0] = 'z'
	println("box:", equal, boxed.(string), string(a))
	a[0] = 'a'
	equal, snapshot = mutateByteString(a)
	println("mutation:", equal, snapshot, string(a))
	a[0] = 'a'
	println("mutation single use:", compareByteStringsAfterStore(a, a))
	a[0] = 'a'
	println("mutation aliased call:", compareByteStringsAfterCall(a, a))
	a[0] = 'a'
	println("mutation cross block:", compareByteStringsAcrossBlock(a, a, false),
		compareByteStringsAcrossBlock(a, a, true))
	a[0] = 'a'
	equal, unequal := compareByteStrings(a, b)
	a[0] = 'z'
	println("after:", equal, unequal, string(a))
	a[0] = 'a'
	compareByteStringFallbacks(a, b, "abc")

	empty := make([]byte, 0, 1)
	emptyWithBacking := []byte{'x'}[:0]
	equal, unequal = compareByteStrings(empty, emptyWithBacking)
	println("empty comparisons:", equal, unequal)

	named := myBytes{'a', 'b', 'c'}
	equal, unequal = compareNamedByteStrings(named, myBytes{'a', 'b', 'c'})
	s, t := convertNamedBytes(named)
	named[0] = 'z'
	println("named:", equal, unequal, s, t)
	runes := []myRune{'a', 0x20ac, 0x1f600}
	s = convertNamedRunes(runes)
	runes[0] = 'z'
	println("named runes:", s)

	data := []byte("prefix-0123456789abcdefghijkl-suffix")
	sep := []byte("0123456789abcdefghijkl")
	println("index:", findBytes(data, sep), findBytes(sep, data))
	sep[0] = 'z'
	println("index absent:", findBytes(data, sep))
}

var byteStringOrderCases = []struct{ a, b []byte }{
	{nil, nil},
	{nil, []byte{}},
	{[]byte{}, nil},
	{nil, []byte{0}},
	{[]byte{0}, nil},
	{[]byte("abc"), []byte("abc")},
	{[]byte("a"), []byte("aa")},
	{[]byte("aa"), []byte("a")},
	{[]byte("abc"), []byte("abd")},
	{[]byte("abc"), []byte("abb")},
	{[]byte{0, 'b'}, []byte{0, 'a'}},
	{[]byte{0, 0x80}, []byte{0, 0xff}},
	{[]byte{0, 0xff}, []byte{0, 0x80}},
	{[]byte{0x7f}, []byte{0x80}},
	{[]byte{0x80}, []byte{0x7f}},
	{[]byte{0xff}, []byte{0xff}},
}

var namedByteStringOrderCases = []struct{ a, b myBytes }{
	{nil, myBytes{}},
	{myBytes{0, 0x80}, myBytes{0, 0xff}},
	{myBytes{0, 0xff}, myBytes{0, 0x80}},
	{myBytes{'a'}, myBytes{'a', 0}},
	{myBytes{'a', 0}, myBytes{'a'}},
}

func testByteSliceStringOrder(op string, compare func(a, b []byte) bool,
	escape func(a, b []byte) (bool, string), mutate func(a, b []byte, c byte) bool,
	named func(a, b myBytes) bool) {
	for i, tc := range byteStringOrderCases {
		println("order:", op, i, compare(tc.a, tc.b))
	}
	for _, c := range []byte{'a', 'm', 'z'} {
		a := []byte("mbc")
		b := []byte{c, 'b', 'c'}
		result, snapshot := escape(a, b)
		a[0] = 'x'
		println("order escape:", op, c, result, snapshot, string(a))
		a[0] = 'm'
		result = mutate(a, a, c)
		println("order mutation:", op, c, result, string(a))
	}
	for _, tc := range namedByteStringOrderCases {
		println("named order:", op, named(tc.a, tc.b))
	}
}

//go:noinline
func reuseLessByteStrings(a, b []byte) (bool, bool) {
	s := string(a)
	t := string(b)
	return s < t, t < s
}

//go:noinline
func lessByteStrings(a, b []byte) bool {
	return string(a) < string(b)
}

//go:noinline
func escapeLessByteString(a, b []byte) (bool, string) {
	s := string(a)
	t := string(b)
	return s < t, s
}

//go:noinline
func mutateLessByteString(a, b []byte, c byte) bool {
	s := string(a)
	b[0] = c
	return s < string(b)
}

//go:noinline
func namedLessByteStrings(a, b myBytes) bool {
	return myString(a) < myString(b)
}

//go:noinline
func lessEqualByteStrings(a, b []byte) bool {
	return string(a) <= string(b)
}

//go:noinline
func escapeLessEqualByteString(a, b []byte) (bool, string) {
	s := string(a)
	t := string(b)
	return s <= t, s
}

//go:noinline
func mutateLessEqualByteString(a, b []byte, c byte) bool {
	s := string(a)
	b[0] = c
	return s <= string(b)
}

//go:noinline
func namedLessEqualByteStrings(a, b myBytes) bool {
	return myString(a) <= myString(b)
}

//go:noinline
func greaterByteStrings(a, b []byte) bool {
	return string(a) > string(b)
}

//go:noinline
func escapeGreaterByteString(a, b []byte) (bool, string) {
	s := string(a)
	t := string(b)
	return s > t, s
}

//go:noinline
func mutateGreaterByteString(a, b []byte, c byte) bool {
	s := string(a)
	b[0] = c
	return s > string(b)
}

//go:noinline
func namedGreaterByteStrings(a, b myBytes) bool {
	return myString(a) > myString(b)
}

//go:noinline
func greaterEqualByteStrings(a, b []byte) bool {
	return string(a) >= string(b)
}

//go:noinline
func escapeGreaterEqualByteString(a, b []byte) (bool, string) {
	s := string(a)
	t := string(b)
	return s >= t, s
}

//go:noinline
func mutateGreaterEqualByteString(a, b []byte, c byte) bool {
	s := string(a)
	b[0] = c
	return s >= string(b)
}

//go:noinline
func namedGreaterEqualByteStrings(a, b myBytes) bool {
	return myString(a) >= myString(b)
}

func main() {
	testRangeString()
	testStringToRunes()
	testRunesToString([]rune{97, 98, 99, 252, 162, 8364, 66376, 176, 120})
	testByteSliceStringCompareNil()
	testByteSliceStringCompareEmpty()
	testByteSliceStringComparisons()
	testByteSliceStringOrder("<", lessByteStrings, escapeLessByteString,
		mutateLessByteString, namedLessByteStrings)
	lt, gt := reuseLessByteStrings([]byte("abc"), []byte("abd"))
	println("order reuse:", lt, gt)
	testByteSliceStringOrder("<=", lessEqualByteStrings, escapeLessEqualByteString,
		mutateLessEqualByteString, namedLessEqualByteStrings)
	testByteSliceStringOrder(">", greaterByteStrings, escapeGreaterByteString,
		mutateGreaterByteString, namedGreaterByteStrings)
	testByteSliceStringOrder(">=", greaterEqualByteStrings, escapeGreaterEqualByteString,
		mutateGreaterEqualByteString, namedGreaterEqualByteStrings)
	var _ = len([]byte(myString("foobar"))) // issue 1246
}
