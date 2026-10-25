//go:build tinygo && !avr

package main

import (
	"runtime"

	"github.com/tinygo-org/tinygo/tools/qbs/dev"
)

func init() {
	extraTests = append(extraTests,
		dev.Test{Name: "map-churn", Run: mapChurn},
		dev.Test{Name: "fragmentation", Run: fragmentation},
	)
}

//go:noinline
func mapTest(n int) bool {
	m := make(map[int]*[8]uint32)
	for i := 0; i < n; i++ {
		v := new([8]uint32)
		v[0] = uint32(i)
		v[7] = uint32(i) * 3
		m[i] = v
	}
	for i := 0; i < n; i += 2 {
		delete(m, i)
	}
	for i := 0; i < n; i++ {
		v, ok := m[i]
		if i%2 == 0 {
			if ok {
				return false
			}
		} else if !ok || v[0] != uint32(i) || v[7] != uint32(i)*3 {
			return false
		}
	}
	for i := 0; i < n; i += 2 {
		m[i] = new([8]uint32)
	}
	return len(m) == n
}

func mapChurn(t *dev.T) {
	b := budget(t, 6)
	if b == 0 {
		return
	}
	n := b / 100
	if n < 4 {
		t.Skip("heap too small")
		return
	}
	noGrowth(t, "map", b, func() bool { return mapTest(n) })
}

//go:noinline
func fragTest(count int) bool {
	blocks := make([][]byte, count)
	size := func(i, round int) int { return 16 + (i*37+round*11)%240 }
	fill := func(i, round int) {
		b := make([]byte, size(i, round))
		for j := range b {
			b[j] = byte(i*5 + round + j)
		}
		blocks[i] = b
	}
	for i := range blocks {
		fill(i, 0)
	}
	for i := 0; i < count; i += 2 {
		blocks[i] = nil
	}
	runtime.GC()
	for i := 0; i < count; i += 2 {
		fill(i, 1)
	}
	for i := range blocks {
		round := 0
		if i%2 == 0 {
			round = 1
		}
		b := blocks[i]
		if len(b) != size(i, round) {
			return false
		}
		for j := range b {
			if b[j] != byte(i*5+round+j) {
				return false
			}
		}
	}
	return true
}

func fragmentation(t *dev.T) {
	b := budget(t, 4)
	if b == 0 {
		return
	}
	count := b / (136 + 24 + 8)
	if count < 8 {
		t.Skip("heap too small")
		return
	}
	noGrowth(t, "fragmentation blocks", b, func() bool { return fragTest(count) })
}
