//go:build tinygo && gc.conservative

package stringcompare

import (
	"runtime"
	"strconv"
	"testing"
)

func checkAllocs(t *testing.T, compare func([]byte, string) bool, n int, wantAllocs uint64) {
	t.Helper()
	a := make([]byte, n)
	for i := range a {
		a[i] = 'a'
	}
	s := string(a)
	const runs = 100
	wantMatches := 0
	for i := 0; i < 2; i++ {
		if n > 0 {
			a[0] = byte('a' + i)
		}
		if compare(a, s) {
			wantMatches += runs / 2
		}
	}
	runtime.GC()
	var before, after runtime.MemStats
	runtime.ReadMemStats(&before)
	matches := 0
	for i := 0; i < runs; i++ {
		if n > 0 {
			a[0] = byte('a' + i&1)
		}
		if compare(a, s) {
			matches++
		}
	}
	runtime.ReadMemStats(&after)
	if matches != wantMatches {
		t.Errorf("got %d matches, want %d", matches, wantMatches)
	}
	if got := after.Mallocs - before.Mallocs; got != runs*wantAllocs {
		t.Errorf("got %d allocations, want %d", got, runs*wantAllocs)
	}
	if got, want := after.TotalAlloc-before.TotalAlloc, runs*wantAllocs*uint64(n); got != want {
		t.Errorf("got %d allocated bytes, want %d", got, want)
	}
}

func TestComparisonAllocs(t *testing.T) {
	for _, tc := range comparisons {
		for _, n := range []int{0, 3, 64, 4096} {
			t.Run(tc.name+"/"+strconv.Itoa(n), func(t *testing.T) {
				checkAllocs(t, tc.compare, n, 0)
			})
		}
	}
	for _, tc := range []struct {
		name    string
		compare func([]byte, string) bool
	}{
		{"reused", reused},
		{"afterCall", afterCall},
	} {
		t.Run(tc.name, func(t *testing.T) {
			checkAllocs(t, tc.compare, 64, 1)
		})
	}
}
