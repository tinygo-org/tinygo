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

//go:noinline
func mixedByteString(a []byte, s string) [6]bool {
	return [6]bool{string(a) == s, string(a) != s, string(a) < s,
		string(a) <= s, string(a) > s, string(a) >= s}
}

//go:noinline
func mixedStringByte(s string, a []byte) [6]bool {
	return [6]bool{s == string(a), s != string(a), s < string(a),
		s <= string(a), s > string(a), s >= string(a)}
}

//go:noinline
func mixedByteLiteral(a []byte) [6]bool {
	return [6]bool{string(a) == "abc", string(a) != "abc", string(a) < "abc",
		string(a) <= "abc", string(a) > "abc", string(a) >= "abc"}
}

//go:noinline
func mixedLiteralByte(a []byte) [6]bool {
	return [6]bool{"abc" == string(a), "abc" != string(a), "abc" < string(a),
		"abc" <= string(a), "abc" > string(a), "abc" >= string(a)}
}

//go:noinline
func mixedNamedByteString(a myBytes, s myString) [6]bool {
	return [6]bool{myString(a) == s, myString(a) != s, myString(a) < s,
		myString(a) <= s, myString(a) > s, myString(a) >= s}
}

//go:noinline
func mixedNamedStringByte(s myString, a myBytes) [6]bool {
	return [6]bool{s == myString(a), s != myString(a), s < myString(a),
		s <= myString(a), s > myString(a), s >= myString(a)}
}

//go:noinline
func mixedStringReference(s, t string) [6]bool {
	return [6]bool{s == t, s != t, s < t, s <= t, s > t, s >= t}
}

//go:noinline
func mixedByteStringReuse(a []byte, s string) (bool, bool) {
	t := string(a)
	equal := t == s
	a[0] = 'z'
	return equal, t < s
}

//go:noinline
func mixedByteStringEscape(a []byte, s string) (bool, string) {
	t := string(a)
	return t == s, t
}

//go:noinline
func mixedByteStringStore(a []byte, s string, dst *string) bool {
	t := string(a)
	equal := t == s
	*dst = t
	return equal
}

//go:noinline
func mixedByteStringBox(a []byte, s string) (bool, any) {
	t := string(a)
	return t == s, any(t)
}

//go:noinline
func mixedByteStringMutation(a, b []byte, s string) bool {
	t := string(a)
	b[0] = 'z'
	return t == s
}

//go:noinline
func mixedByteStringCall(a, b []byte, s string) bool {
	t := string(a)
	mutateByteSlice(b)
	return s < t
}

//go:noinline
func mixedByteStringAcrossBlock(a, b []byte, s string, mutate bool) bool {
	t := string(a)
	if mutate {
		b[0] = 'z'
	}
	return t == s
}

//go:noinline
func mixedByteStringAfter(a []byte, s string) bool {
	result := string(a) == s
	a[0] = 'z'
	return result
}

var mixedEvaluationOrder int

//go:noinline
func mixedBytesOperand(a []byte) []byte {
	mixedEvaluationOrder = mixedEvaluationOrder*10 + 1
	return a
}

//go:noinline
func mixedStringOperand(s string) string {
	mixedEvaluationOrder = mixedEvaluationOrder*10 + 2
	return s
}

//go:noinline
func mixedByteStringEvaluation(a []byte, s string) bool {
	return string(mixedBytesOperand(a)) == mixedStringOperand(s)
}

//go:noinline
func mixedStringByteEvaluation(s string, a []byte) bool {
	return mixedStringOperand(s) == string(mixedBytesOperand(a))
}

func testMixedByteStringComparisons() {
	cases := append(byteStringOrderCases, struct{ a, b []byte }{
		make([]byte, 0, 1), []byte{'x'}[:0],
	})
	for _, tc := range cases {
		s, t := string(tc.a), string(tc.b)
		if mixedByteString(tc.a, t) != mixedStringReference(s, t) ||
			mixedStringByte(t, tc.a) != mixedStringReference(t, s) {
			panic("mixed string comparison")
		}
		if mixedByteLiteral(tc.a) != mixedStringReference(s, "abc") ||
			mixedLiteralByte(tc.a) != mixedStringReference("abc", s) {
			panic("mixed literal comparison")
		}
		var named myBytes
		for _, c := range tc.a {
			named = append(named, myByte(c))
		}
		if mixedNamedByteString(named, myString(t)) != mixedStringReference(s, t) ||
			mixedNamedStringByte(myString(t), named) != mixedStringReference(t, s) {
			panic("mixed named string comparison")
		}
	}

	a := []byte("abc")
	equal, less := mixedByteStringReuse(a, "abd")
	if equal || !less {
		panic("mixed reused snapshot")
	}
	a[0] = 'a'
	equal, snapshot := mixedByteStringEscape(a, "abc")
	a[0] = 'z'
	if !equal || snapshot != "abc" {
		panic("mixed escaping snapshot")
	}
	a[0] = 'a'
	equal = mixedByteStringStore(a, "abc", &snapshot)
	a[0] = 'z'
	if !equal || snapshot != "abc" {
		panic("mixed stored snapshot")
	}
	a[0] = 'a'
	equal, boxed := mixedByteStringBox(a, "abc")
	a[0] = 'z'
	if !equal || boxed.(string) != "abc" {
		panic("mixed boxed snapshot")
	}
	a[0] = 'a'
	if !mixedByteStringMutation(a, a, "abc") {
		panic("mixed aliased mutation")
	}
	a[0] = 'a'
	if mixedByteStringCall(a, a, "mbc") {
		panic("mixed aliased call")
	}
	for _, mutate := range []bool{false, true} {
		a[0] = 'a'
		if !mixedByteStringAcrossBlock(a, a, "abc", mutate) {
			panic("mixed cross block snapshot")
		}
	}
	a[0] = 'a'
	if !mixedByteStringAfter(a, "abc") || a[0] != 'z' {
		panic("mixed mutation after comparison")
	}
	a[0] = 'a'
	mixedEvaluationOrder = 0
	if !mixedByteStringEvaluation(a, "abc") || mixedEvaluationOrder != 12 {
		panic("mixed left evaluation order")
	}
	mixedEvaluationOrder = 0
	if !mixedStringByteEvaluation("abc", a) || mixedEvaluationOrder != 21 {
		panic("mixed right evaluation order")
	}
	println("mixed comparisons:", len(cases)*6*6, "snapshot and evaluation guards passed")
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
	testMixedByteStringComparisons()
	var _ = len([]byte(myString("foobar"))) // issue 1246
}
