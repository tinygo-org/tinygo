package main

func someString() string {
	return "foo"
}

func zeroLengthString() string {
	return ""
}

func stringLen(s string) int {
	return len(s)
}

func stringIndex(s string, index int) byte {
	return s[index]
}

func stringCompareEqual(s1, s2 string) bool {
	return s1 == s2
}

func stringCompareUnequal(s1, s2 string) bool {
	return s1 != s2
}

func byteSliceStringCompareEqual(s1, s2 []byte) bool {
	return string(s1) == string(s2)
}

func byteSliceStringCompareUnequal(s1, s2 []byte) bool {
	return string(s1) != string(s2)
}

func byteSliceStringCompareSideEffects(s1, s2 []byte) bool {
	return string(s1) == string(mutateBytes(s2))
}

func byteSliceStringCompareNil(s []byte) bool {
	var nilSlice []byte
	return string(s) == string(nilSlice)
}

func byteSliceStringCompareUint8(a, b []uint8) bool {
	return string(a) == string(b)
}

func byteSliceStringCompareEmpty(a, b []byte) bool {
	return string(a[:0]) == string(b[:0])
}

func byteSliceStringCompareLocal(a, b []byte) bool {
	s := string(a)
	t := string(b)
	return s == t
}

func byteSliceStringCompareEscape(a, b []byte) (bool, string) {
	s := string(a)
	t := string(b)
	return s == t, s
}

func byteSliceStringCompareStore(a, b []byte, dst *string) bool {
	s := string(a)
	t := string(b)
	equal := s == t
	*dst = s
	return equal
}

func byteSliceStringCompareBox(a, b []byte) (bool, any) {
	s := string(a)
	t := string(b)
	return s == t, any(s)
}

func byteSliceStringCompareMutation(a, b []byte) bool {
	s := string(a)
	mutateBytes(b)
	return s == string(b)
}

func byteSliceStringCompareAfterMutation(a, b []byte) (bool, string) {
	s := string(a)
	t := string(b)
	equal := s == t
	mutateBytes(a)
	return equal, s
}

func byteSliceStringCompareSlices(a, b []byte) bool {
	return string(a[:2]) == string(b[:2])
}

func byteSliceStringCompareLiteral(a []byte) bool {
	return string(a) == "abc"
}

func byteSliceStringCompareOrdered(a, b []byte) bool {
	return string(a) < string(b)
}

type namedByte byte
type namedBytes []namedByte
type namedRune rune
type namedString string

func namedByteSliceStringCompare(a, b namedBytes) bool {
	return string(a) == string(b)
}

func byteSliceNamedStringCompare(a, b []byte) bool {
	return namedString(a) != namedString(b)
}

func namedByteSliceToString(a namedBytes) string {
	return string(a)
}

func namedByteSliceToNamedString(a namedBytes) namedString {
	return namedString(a)
}

func namedRuneSliceToString(a []namedRune) string {
	return string(a)
}

//go:noinline
func mutateBytes(s []byte) []byte {
	s[0]++
	return s
}

func stringCompareLarger(s1, s2 string) bool {
	return s1 > s2
}

func stringLookup(s string, x uint8) byte {
	// Test that x is correctly extended to an uint before comparison.
	return s[x]
}
