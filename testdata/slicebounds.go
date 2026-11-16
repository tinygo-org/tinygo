package main

import (
	"reflect"
	"runtime"
)

// This test checks that slicing a slice panics exactly when the Go spec says
// it must, and otherwise gives the right len and cap. The expected output is
// the same for gc and TinyGo.

var sink []byte

//go:noinline
func id(n int) int { return n }

// slice runs f and reports whether it panicked with a runtime.Error.
func slice(f func() []byte) (s []byte, panicked bool) {
	defer func() {
		if r := recover(); r != nil {
			if _, ok := r.(runtime.Error); !ok {
				println("panic is not a runtime.Error")
			}
			panicked = true
		}
	}()
	return f(), false
}

// reflectSlice is like slice, but reflect panics with a string rather than a
// runtime.Error.
func reflectSlice(f func() []byte) (s []byte, panicked bool) {
	defer func() {
		if recover() != nil {
			panicked = true
		}
	}()
	return f(), false
}

var cases, failures int

func check(name string, l, c, low, high, max int, s []byte, panicked bool) {
	cases++
	valid := 0 <= low && low <= high && high <= max && max <= c
	switch {
	case valid == panicked:
		println("wrong panic behavior:", name, "len", l, "cap", c, "low", low, "high", high, "max", max, "panicked", panicked)
		failures++
	case valid && (len(s) != high-low || cap(s) != max-low):
		println("wrong len or cap:", name, "len", l, "cap", c, "low", low, "high", high, "max", max, "got", len(s), cap(s))
		failures++
	}
}

func main() {
	for c := 0; c <= 6; c++ {
		for l := 0; l <= c; l++ {
			// The slice has spare capacity after its length.
			b := make([]byte, c)[:l]
			for low := -1; low <= c+1; low++ {
				for high := -1; high <= c+1; high++ {
					for max := -1; max <= c+1; max++ {
						lo, hi, mx := id(low), id(high), id(max)
						s, p := slice(func() []byte { return b[lo:] })
						check("b[low:]", l, c, low, l, c, s, p)
						s, p = slice(func() []byte { return b[:hi] })
						check("b[:high]", l, c, 0, high, c, s, p)
						s, p = slice(func() []byte { return b[lo:hi] })
						check("b[low:high]", l, c, low, high, c, s, p)
						s, p = slice(func() []byte { return b[:hi:mx] })
						check("b[:high:max]", l, c, 0, high, max, s, p)
						s, p = slice(func() []byte { return b[lo:hi:mx] })
						check("b[low:high:max]", l, c, low, high, max, s, p)
					}
				}
			}
		}
	}

	// reflect must agree with the language.
	for c := 0; c <= 5; c++ {
		for l := 0; l <= c; l++ {
			b := make([]byte, c)[:l]
			v := reflect.ValueOf(b)
			arr := reflect.ValueOf(&[5]byte{}).Elem()
			for low := -1; low <= c+1; low++ {
				for high := -1; high <= c+1; high++ {
					lo, hi := id(low), id(high)
					s, p := reflectSlice(func() []byte { return v.Slice(lo, hi).Bytes() })
					check("reflect Slice", l, c, low, high, c, s, p)
					// Arrays can be sliced up to their length, which is 5.
					s, p = reflectSlice(func() []byte { return arr.Slice(lo, hi).Bytes() })
					check("reflect Slice array", 5, 5, low, high, 5, s, p)
					for max := -1; max <= c+1; max++ {
						mx := id(max)
						s, p = reflectSlice(func() []byte { return v.Slice3(lo, hi, mx).Bytes() })
						check("reflect Slice3", l, c, low, high, max, s, p)
						s, p = reflectSlice(func() []byte { return arr.Slice3(lo, hi, mx).Bytes() })
						check("reflect Slice3 array", 5, 5, low, high, max, s, p)
					}
				}
			}
		}
	}

	// Strings can be sliced up to their length.
	str := reflect.ValueOf("abcd")
	for low := -1; low <= 5; low++ {
		for high := -1; high <= 5; high++ {
			lo, hi := id(low), id(high)
			// The capacity of a converted string is unspecified, so pin it.
			s, p := reflectSlice(func() []byte {
				b := []byte(str.Slice(lo, hi).String())
				return b[:len(b):len(b)]
			})
			check("reflect Slice string", 4, 4, low, high, high, s, p)
		}
	}

	// A nil slice has len and cap 0.
	var nilSlice []byte
	for n := -1; n <= 2; n++ {
		n := id(n)
		s, p := slice(func() []byte { return nilSlice[:n] })
		check("nil[:n]", 0, 0, 0, n, 0, s, p)
		s, p = slice(func() []byte { return nilSlice[n:] })
		check("nil[n:]", 0, 0, n, 0, 0, s, p)
	}

	// Reslicing after a length guard must only panic when n is past the cap.
	for c := 0; c <= 6; c++ {
		for l := 0; l <= c; l++ {
			b := make([]byte, c)[:l]
			for n := 0; n <= c+1; n++ {
				n := id(n)
				s, p := slice(func() []byte {
					if len(b) >= n {
						return b[:n]
					}
					return b[:n:n]
				})
				max := n
				if l >= n {
					max = c
				}
				check("guarded", l, c, 0, n, max, s, p)
			}
		}
	}

	// Consume a slice in steps, as hash functions do.
	for l := 0; l <= 20; l++ {
		p := make([]byte, l, l+3)
		steps := 0
		for len(p) >= 4 {
			sink = p[:4]
			p = p[4:]
			steps++
		}
		if steps != l/4 || len(p) != l%4 {
			println("wrong step loop:", l, steps, len(p))
			failures++
		}
	}

	println("cases:", cases, "failures:", failures)
}
