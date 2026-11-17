//go:build tinygo

package escapeallocs

import "testing"

var (
	sinkInt int
	input   = []uint16{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	other   = make([]byte, 16)
)

func checkNoAllocs(t *testing.T, fn func()) {
	t.Helper()
	if allocs := testing.AllocsPerRun(100, fn); allocs != 0 {
		t.Errorf("got %v allocations, want 0", allocs)
	}
}

// appendAll carries buf around a loop and returns it.
//
//go:noinline
func appendAll(buf []rune, s []uint16) []rune {
	for _, c := range s {
		buf = append(buf, rune(c))
	}
	return buf
}

// A buffer that is only passed to a function that appends to it in a loop.
func TestAppendInCallee(t *testing.T) {
	checkNoAllocs(t, func() {
		buf := make([]rune, 0, 64)
		sinkInt = len(appendAll(buf, input))
	})
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

type uint128 struct{ hi, lo uint64 }

func (u *uint128) halves() [2]*uint64 {
	return [2]*uint64{&u.hi, &u.lo}
}

type addr struct {
	addr uint128
	z    int
}

// v6 takes the address of its receiver copy and picks one of two pointers into
// it with a run time index, like net/netip.Addr.v6.
//
//go:noinline
func (a addr) v6(i uint8) uint8 {
	return uint8(*(a.addr.halves()[(i/8)%2]) >> ((7 - i%8) * 8))
}

// A pointer is stored into a temporary array on the stack and loaded again.
func TestPointerThroughLocalArray(t *testing.T) {
	checkNoAllocs(t, func() {
		a := addr{addr: uint128{hi: 0x0102030405060708, lo: 0x1112131415161718}}
		sinkInt = int(a.v6(3)) + int(a.v6(12))
	})
}

// Objects allocated in a loop and kept in a local array across iterations must
// stay on the heap, one allocation per iteration, and keep their values.
func TestAllocInLoopKeptInArray(t *testing.T) {
	var sum int
	allocs := testing.AllocsPerRun(100, func() {
		var keep [2]*[16]byte
		sum = 0
		for i := 0; i < 6; i++ {
			b := new([16]byte)
			b[0] = byte(i + 1)
			keep[i&1] = b
			if other := keep[(i+1)&1]; other != nil {
				sum += int(other[0])
			}
		}
	})
	if allocs != 6 {
		t.Errorf("got %v allocations, want 6", allocs)
	}
	if sum != 1+2+3+4+5 {
		t.Errorf("got sum %d, want %d", sum, 1+2+3+4+5)
	}
}
