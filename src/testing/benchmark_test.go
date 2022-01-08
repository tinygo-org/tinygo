// Copyright 2011 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package testing_test

import (
	"testing"
)

var buf = make([]byte, 13579)

func NonASCII(b []byte, i int, offset int) int {
	for i = offset; i < len(b)+offset; i++ {
		if b[i%len(b)] >= 0x80 {
			break
		}
	}
	return i
}

func BenchmarkFastNonASCII(b *testing.B) {
	var val int
	for i := 0; i < b.N; i++ {
		val += NonASCII(buf, 0, 0)
	}
}

func BenchmarkSlowNonASCII(b *testing.B) {
	var val int
	for i := 0; i < b.N; i++ {
		val += NonASCII(buf, 0, 0)
		val += NonASCII(buf, 0, 1)
	}
}

// TestBenchmark simply uses Benchmark twice and makes sure it does not crash.
func TestBenchmark(t *testing.T) {
	// FIXME: reduce runtime from the current 3 seconds.
	rslow := testing.Benchmark(BenchmarkSlowNonASCII)
	rfast := testing.Benchmark(BenchmarkFastNonASCII)
	tslow := rslow.NsPerOp()
	tfast := rfast.NsPerOp()

	// Be exceedingly forgiving; do not fail even if system gets busy.
	speedup := float64(tslow) / float64(tfast)
	if speedup < 0.3 {
		t.Errorf("Expected speedup >= 0.3, got %f", speedup)
	}
}

func BenchmarkSub(b *testing.B) {
	b.Run("Fast", func(b *testing.B) { BenchmarkFastNonASCII(b) })
	b.Run("Slow", func(b *testing.B) { BenchmarkSlowNonASCII(b) })
}
