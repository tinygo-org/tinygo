//go:build tinygo

package escapeallocs

import "testing"

var (
	sinkInt int
	other   = make([]byte, 16)
)

func checkNoAllocs(t *testing.T, fn func()) {
	t.Helper()
	if allocs := testing.AllocsPerRun(100, fn); allocs != 0 {
		t.Errorf("got %v allocations, want 0", allocs)
	}
}

// A buffer that is allocated once and merged with another slice inside a loop.
func TestMergeInLoopBody(t *testing.T) {
	checkNoAllocs(t, func() {
		buf := make([]byte, 16)
		sum := 0
		for i := 0; i < 8; i++ {
			p := buf
			if i&1 == 0 {
				p = other
			}
			sum += int(p[i])
		}
		sinkInt = sum
	})
}

// A buffer that is allocated in a loop and kept for the next iteration must
// still be allocated on the heap, so there is exactly one allocation per
// iteration.
func TestAllocInLoopKeptAcrossIterations(t *testing.T) {
	allocs := testing.AllocsPerRun(100, func() {
		var prev []byte
		for i := 0; i < 4; i++ {
			buf := make([]byte, 16)
			buf[0] = byte(i)
			if prev != nil {
				sinkInt += int(prev[0])
			}
			prev = buf
		}
	})
	if allocs != 4 {
		t.Errorf("got %v allocations, want 4", allocs)
	}
}
