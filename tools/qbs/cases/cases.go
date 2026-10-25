// Package cases has deterministic computations that run on the device and on
// the host. The host compares the two results.
package cases

import "strconv"

// Case is one computation. Fn returns a short string that describes the result.
type Case struct {
	Name string
	Fn   func() string
	// ULP is the allowed difference in units in the last place when Fn
	// returns space separated hex float bits. Zero means an exact match.
	ULP uint64
	// Heap is the number of bytes of heap the case needs. The device skips
	// the case when less is free.
	Heap int
}

var all []Case

func register(name string, ulp uint64, fn func() string) {
	all = append(all, Case{Name: name, Fn: fn, ULP: ulp})
}

// needsHeap records how much heap the case registered last needs.
func needsHeap(bytes int) {
	all[len(all)-1].Heap = bytes
}

// All returns the cases that are built for this target.
func All() []Case {
	return all
}

// rng is a xorshift32 generator.
type rng uint32

func (r *rng) next() uint32 {
	x := uint32(*r)
	x ^= x << 13
	x ^= x >> 17
	x ^= x << 5
	*r = rng(x)
	return x
}

func randBytes(n int, seed uint32) []byte {
	r := rng(seed)
	b := make([]byte, n)
	for i := range b {
		b[i] = byte(r.next() >> 11)
	}
	return b
}

func hex64(v uint64) string {
	return strconv.FormatUint(v, 16)
}

// fnv64 hashes s so that long results stay short.
func fnv64(s string) uint64 {
	h := uint64(14695981039346656037)
	for i := 0; i < len(s); i++ {
		h ^= uint64(s[i])
		h *= 1099511628211
	}
	return h
}
