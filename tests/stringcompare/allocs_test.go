//go:build tinygo

package stringcompare

import (
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
	var want [2]bool
	for i := range want {
		if n > 0 {
			a[0] = byte('a' + i)
		}
		want[i] = compare(a, s)
	}
	i, mismatches := 0, 0
	allocs := testing.AllocsPerRun(100, func() {
		if n > 0 {
			a[0] = byte('a' + i&1)
		}
		if compare(a, s) != want[i&1] {
			mismatches++
		}
		i++
	})
	if mismatches != 0 {
		t.Errorf("got %d wrong results", mismatches)
	}
	if allocs != float64(wantAllocs) {
		t.Errorf("got %v allocations, want %d", allocs, wantAllocs)
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
