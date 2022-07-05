package main

import (
	"hash/maphash"
	"strconv"
	"testing"
)

var buf [8192]byte

func BenchmarkMaphash(b *testing.B) {
	var h maphash.Hash
	benchmarkHash(b, "maphash", h)
}

func benchmarkHash(b *testing.B, str string, h maphash.Hash) {
	var sizes = []int{1, 2, 3, 4, 5, 10, 20, 30, 40, 50, 60, 70, 80, 90, 100, 1024, 8192}
	for _, n := range sizes {
		b.Run(strconv.Itoa(n), func(b *testing.B) { benchmarkHashn(b, int64(n), h) })
	}
}

var total uint64

func benchmarkHashn(b *testing.B, size int64, h maphash.Hash) {
	b.SetBytes(size)

	sum := make([]byte, 4)

	for i := 0; i < b.N; i++ {
		h.Reset()
		h.Write(buf[:size])
		sum = h.Sum(sum[:0])
		total += uint64(sum[0])
	}
}
