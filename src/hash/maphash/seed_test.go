// This file has no upstream equivalent. It covers a TinyGo defect where
// runtime.comparablehash ignored the seed for a value with no bytes in it.

package maphash

import "testing"

// seedsPerValue is the number of random seeds each value is hashed under. Two
// of them collide by chance with probability 2^-62, so this is not flaky.
const seedsPerValue = 32

func TestComparableSeedDependent(t *testing.T) {
	t.Run("nil interface", func(t *testing.T) { testSeedDependent(t, any(nil)) })
	t.Run("empty struct", func(t *testing.T) { testSeedDependent(t, struct{}{}) })
	t.Run("empty array", func(t *testing.T) { testSeedDependent(t, [0]int{}) })
	t.Run("array of empty struct", func(t *testing.T) { testSeedDependent(t, [2]struct{}{}) })
	t.Run("struct of empty struct", func(t *testing.T) { testSeedDependent(t, struct{ a struct{} }{}) })
	t.Run("struct of empty array", func(t *testing.T) { testSeedDependent(t, struct{ a [0]int }{}) })
}

func testSeedDependent[T comparable](t *testing.T, v T) {
	t.Helper()

	sums := make(map[uint64]Seed, seedsPerValue)
	for range seedsPerValue {
		seed := MakeSeed()
		sum := Comparable(seed, v)
		if got := Comparable(seed, v); got != sum {
			t.Fatalf("Comparable(%v, %#v) = %#016x, then %#016x, want one value", seed, v, sum, got)
		}
		if prev, ok := sums[sum]; ok {
			t.Fatalf("Comparable(%v, %#v) and Comparable(%v, %#v) are both %#016x, want the seed to change the result", prev, v, seed, v, sum)
		}
		sums[sum] = seed
	}
}

// TestWriteComparableKeepsSeed makes sure that a value with no bytes in it does
// not set the state of the hash to 0 and discard the seed.
func TestWriteComparableKeepsSeed(t *testing.T) {
	var h1, h2 Hash
	h1.SetSeed(MakeSeed())
	h2.SetSeed(MakeSeed())

	WriteComparable(&h1, struct{}{})
	WriteComparable(&h2, struct{}{})
	if h1.Sum64() == h2.Sum64() {
		t.Error("WriteComparable of an empty struct gives one sum for two seeds")
	}

	h1.WriteString("abc")
	h2.WriteString("abc")
	if h1.Sum64() == h2.Sum64() {
		t.Error("a write after WriteComparable of an empty struct gives one sum for two seeds")
	}
}
