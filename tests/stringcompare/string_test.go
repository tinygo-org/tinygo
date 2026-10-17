package stringcompare

import (
	"strconv"
	"testing"
)

//go:noinline
func equal(a []byte, s string) bool { return string(a) == s }

//go:noinline
func notEqual(a []byte, s string) bool { return string(a) != s }

//go:noinline
func less(a []byte, s string) bool { return string(a) < s }

//go:noinline
func lessEqual(a []byte, s string) bool { return string(a) <= s }

//go:noinline
func greater(a []byte, s string) bool { return string(a) > s }

//go:noinline
func greaterEqual(a []byte, s string) bool { return string(a) >= s }

//go:noinline
func reverseEqual(a []byte, s string) bool { return s == string(a) }

//go:noinline
func reverseNotEqual(a []byte, s string) bool { return s != string(a) }

//go:noinline
func reverseLess(a []byte, s string) bool { return s < string(a) }

//go:noinline
func reverseLessEqual(a []byte, s string) bool { return s <= string(a) }

//go:noinline
func reverseGreater(a []byte, s string) bool { return s > string(a) }

//go:noinline
func reverseGreaterEqual(a []byte, s string) bool { return s >= string(a) }

var comparisons = []struct {
	name    string
	compare func([]byte, string) bool
}{
	{"equal", equal},
	{"notEqual", notEqual},
	{"less", less},
	{"lessEqual", lessEqual},
	{"greater", greater},
	{"greaterEqual", greaterEqual},
	{"reverseEqual", reverseEqual},
	{"reverseNotEqual", reverseNotEqual},
	{"reverseLess", reverseLess},
	{"reverseLessEqual", reverseLessEqual},
	{"reverseGreater", reverseGreater},
	{"reverseGreaterEqual", reverseGreaterEqual},
}

func byteOrder(a, b []byte) int {
	for i := 0; i < len(a) && i < len(b); i++ {
		if a[i] < b[i] {
			return -1
		}
		if a[i] > b[i] {
			return 1
		}
	}
	if len(a) < len(b) {
		return -1
	}
	if len(a) > len(b) {
		return 1
	}
	return 0
}

func TestBoundaries(t *testing.T) {
	for _, n := range []int{0, 1, 2, 3, 7, 8, 15, 16, 31, 32, 33, 63, 64, 65, 127, 128, 129, 255, 256, 257, 4095, 4096} {
		backing := make([]byte, n+8)
		other := make([]byte, n+1)
		for offset := 0; offset < 8; offset++ {
			a := backing[offset : offset+n]
			for i := range a {
				a[i] = byte(i*97 + offset*19)
			}
			for _, mismatch := range []string{"equal", "first", "last", "shorter", "longer"} {
				b := other[:n]
				copy(b, a)
				switch mismatch {
				case "first":
					if n > 0 {
						b[0] ^= 0x80
					}
				case "last":
					if n > 0 {
						b[n-1] ^= 0x80
					}
				case "shorter":
					if n > 0 {
						b = b[:n-1]
					}
				case "longer":
					b = other[:n+1]
					b[n] = 0
				}
				order := byteOrder(a, b)
				want := [12]bool{
					order == 0, order != 0, order < 0, order <= 0, order > 0, order >= 0,
					order == 0, order != 0, order > 0, order >= 0, order < 0, order <= 0,
				}
				s := string(b)
				for i, tc := range comparisons {
					if got := tc.compare(a, s); got != want[i] {
						t.Fatalf("length %d offset %d %s %s: got %v, want %v", n, offset, mismatch, tc.name, got, want[i])
					}
				}
			}
		}
	}
}

//go:noinline
func loadRight(a []byte, s *string) bool { return string(a) == *s }

//go:noinline
func loadLeft(a []byte, s *string) bool { return *s == string(a) }

//go:noinline
func sliceRight(a []byte, s string, n int) bool { return string(a) == s[:n] }

//go:noinline
func sliceLeft(a []byte, s string, n int) bool { return s[:n] == string(a) }

func mustPanic(t *testing.T, f func()) {
	t.Helper()
	defer func() {
		if recover() == nil {
			t.Error("expected a panic")
		}
	}()
	f()
}

func TestFallbacks(t *testing.T) {
	a := []byte("abc")
	s := "abc"
	if !loadRight(a, &s) || !loadLeft(a, &s) ||
		!sliceRight(a, s, 3) || !sliceLeft(a, s, 3) {
		t.Fatal("load or slice comparison failed")
	}
	for _, tc := range []struct {
		name string
		call func()
	}{
		{"nilRight", func() { loadRight(a, nil) }},
		{"nilLeft", func() { loadLeft(a, nil) }},
		{"boundsRight", func() { sliceRight(a, s, 4) }},
		{"boundsLeft", func() { sliceLeft(a, s, 4) }},
		{"negativeRight", func() { sliceRight(a, s, -1) }},
		{"negativeLeft", func() { sliceLeft(a, s, -1) }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			mustPanic(t, tc.call)
		})
	}
}

//go:noinline
func phiSnapshot(a []byte, alternate string, choose bool) bool {
	s := string(a)
	if choose {
		s = alternate
	}
	a[0] = 'z'
	return s == "abc"
}

func TestPhiSnapshot(t *testing.T) {
	a := []byte("abc")
	if !phiSnapshot(a, "xyz", false) || a[0] != 'z' {
		t.Fatal("phi did not preserve the converted string")
	}
	a[0] = 'a'
	if phiSnapshot(a, "xyz", true) || a[0] != 'z' {
		t.Fatal("phi did not select the alternate string")
	}
}

//go:noinline
func reused(a []byte, s string) bool {
	t := string(a)
	return t == s || t < s
}

//go:noinline
func mutate(a []byte) { a[0] = 'z' }

//go:noinline
func afterCall(a []byte, s string) bool {
	t := string(a)
	mutate(a)
	return t == s
}

func TestCallSnapshot(t *testing.T) {
	a := []byte("abc")
	if !afterCall(a, "abc") || a[0] != 'z' {
		t.Fatal("call did not preserve the converted string")
	}
}

func BenchmarkComparison(b *testing.B) {
	for _, tc := range []struct {
		name    string
		compare func([]byte, string) bool
	}{
		{"equal", equal},
		{"less", less},
		{"reused", reused},
		{"afterCall", afterCall},
	} {
		for _, n := range []int{3, 64, 4096} {
			for _, mismatch := range []string{"equal", "first", "last", "length"} {
				b.Run(tc.name+"/"+strconv.Itoa(n)+"/"+mismatch, func(b *testing.B) {
					a := make([]byte, n)
					other := make([]byte, n)
					for i := range a {
						a[i], other[i] = 'a', 'a'
					}
					var strings [2]string
					for i := range strings {
						other[0] = byte('a' + i)
						switch mismatch {
						case "first":
							other[0] = 'z'
						case "last":
							other[n-1] = 'z'
						}
						strings[i] = string(other)
						if mismatch == "length" {
							strings[i] += "x"
						}
					}
					want := 0
					for i, s := range strings {
						a[0] = byte('a' + i)
						if tc.compare(a, s) {
							want += (b.N + 1 - i) / 2
						}
					}
					b.ReportAllocs()
					b.ResetTimer()
					matches := 0
					for i := 0; i < b.N; i++ {
						a[0] = byte('a' + i&1)
						if tc.compare(a, strings[i&1]) {
							matches++
						}
					}
					b.StopTimer()
					if matches != want {
						b.Fatalf("got %d matches, want %d", matches, want)
					}
				})
			}
		}
	}
}
