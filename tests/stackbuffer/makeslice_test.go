//go:build tinygo

package stackbuffer

import (
	"runtime"
	"testing"
)

var (
	globalBytes  []byte
	globalSlices [][]byte
	sinkPointers []*int
)

// size hides the length from the compiler, so the make has no constant size.
//
//go:noinline
func size(n int) int { return n }

//go:noinline
func fillAndSum(n int) int {
	b := make([]byte, size(n))
	for i := range b {
		b[i] = byte(i + 1)
	}
	sum := 0
	for _, c := range b {
		sum += int(c)
	}
	return sum
}

func TestMakeAllocations(t *testing.T) {
	for _, n := range []int{0, 1, 5, 31, 32, 33, 100} {
		want := n * (n + 1) / 2
		if n > 255 {
			want = 0
		}
		wantAllocs := 0.0
		if n > 32 {
			wantAllocs = 1
		}
		var got int
		allocs := testing.AllocsPerRun(50, func() { got = fillAndSum(n) })
		if got != want {
			t.Errorf("length %d: got %d, want %d", n, got, want)
		}
		if allocs != wantAllocs {
			t.Errorf("length %d: got %v allocations, want %v", n, allocs, wantAllocs)
		}
	}
}

//go:noinline
func dirtyStack() {
	var pad [512]byte
	for i := range pad {
		pad[i] = 0xAA
	}
	sinkInt += int(pad[len(pad)-1])
}

//go:noinline
func zeroedBytes(n int) bool {
	b := make([]byte, size(n))
	for _, c := range b {
		if c != 0 {
			return false
		}
	}
	return len(b) == n && cap(b) == n
}

//go:noinline
func zeroedInts(n int) bool {
	s := make([]int, size(n))
	for _, v := range s {
		if v != 0 {
			return false
		}
	}
	return len(s) == n
}

//go:noinline
func zeroedStrings(n int) bool {
	s := make([]string, size(n))
	for _, v := range s {
		if v != "" {
			return false
		}
	}
	return len(s) == n
}

// The memory is zero even where the stack held other values before.
func TestMakeIsZeroed(t *testing.T) {
	for n := 0; n <= 40; n++ {
		dirtyStack()
		if !zeroedBytes(n) {
			t.Errorf("[]byte of length %d is not zero", n)
		}
		dirtyStack()
		if !zeroedInts(n / 4) {
			t.Errorf("[]int of length %d is not zero", n/4)
		}
		dirtyStack()
		if !zeroedStrings(n / 8) {
			t.Errorf("[]string of length %d is not zero", n/8)
		}
	}
}

// The slice lives on the stack and keeps pointers to heap objects.
func TestMakeKeepsPointersAlive(t *testing.T) {
	p := make([]*int, size(3))
	for i := range p {
		v := new(int)
		*v = 100 + i
		p[i] = v
	}
	for i := 0; i < 2000; i++ {
		sinkPointers = make([]*int, 16)
	}
	runtime.GC()
	for i, v := range p {
		if *v != 100+i {
			t.Errorf("element %d is %d, want %d", i, *v, 100+i)
		}
	}
}

// A fixed length with a capacity that is not constant.
//
//go:noinline
func makeWithCap(n int) []byte {
	b := make([]byte, 3, size(n))
	copy(b, "abc")
	b = append(b, "defghijklmnopqrstuvwxyz0123456789"...)
	return b
}

func TestAppendPastTheBuffer(t *testing.T) {
	for _, n := range []int{3, 5, 32, 33, 64} {
		got := string(makeWithCap(n))
		clobber(8)
		if want := "abcdefghijklmnopqrstuvwxyz0123456789"; got != want {
			t.Errorf("capacity %d: got %q, want %q", n, got, want)
		}
	}
}

//go:noinline
func storeSlice(n int) {
	globalBytes = make([]byte, size(n))
	copy(globalBytes, "stored")
}

//go:noinline
func returnSlice(n int) []byte {
	b := make([]byte, size(n))
	copy(b, "returned")
	return b
}

//go:noinline
func captureSlice(n int) func() string {
	b := make([]byte, size(n))
	copy(b, "captured")
	return func() string { return string(b) }
}

//go:noinline
func sendSlice(n int, ch chan []byte) {
	b := make([]byte, size(n))
	copy(b, "sent")
	ch <- b
}

//go:noinline
func interfaceSlice(n int) {
	b := make([]byte, size(n))
	copy(b, "iface")
	globalAny = b
}

//go:noinline
func appendSlice(n int) {
	b := make([]byte, size(n))
	copy(b, "appended")
	globalSlices = append(globalSlices, b)
}

// A slice that escapes must be on the heap and survive the stack.
func TestEscapingSlices(t *testing.T) {
	storeSlice(10)
	r := returnSlice(10)
	f := captureSlice(10)
	ch := make(chan []byte, 1)
	sendSlice(10, ch)
	interfaceSlice(10)
	appendSlice(10)
	clobber(8)

	check := func(name string, got []byte, want string) {
		t.Helper()
		if len(got) < len(want) || string(got[:len(want)]) != want {
			t.Errorf("%s: got %q, want %q", name, got, want)
		}
	}
	check("global", globalBytes, "stored")
	check("returned", r, "returned")
	if got := f(); got[:8] != "captured" {
		t.Errorf("closure: got %q", got)
	}
	check("channel", <-ch, "sent")
	check("interface", globalAny.([]byte), "iface")
	check("appended", globalSlices[len(globalSlices)-1], "appended")
}

// A slice made in a loop and kept for the next iteration cannot share a buffer.
func TestSlicesKeptFromLoop(t *testing.T) {
	var kept [6][]byte
	allocs := testing.AllocsPerRun(20, func() {
		for i := range kept {
			b := make([]byte, size(8))
			for j := range b {
				b[j] = byte('a' + i)
			}
			kept[i] = b
		}
	})
	if allocs < 6 {
		t.Errorf("got %v allocations, want at least 6", allocs)
	}
	clobber(8)
	for i, b := range kept {
		for _, c := range b {
			if c != byte('a'+i) {
				t.Fatalf("slice %d changed: %q", i, b)
			}
		}
	}
}

func TestPreviousSliceInLoop(t *testing.T) {
	var prev []byte
	changes := 0
	for i := 0; i < 6; i++ {
		cur := make([]byte, size(8))
		for j := range cur {
			cur[j] = byte('a' + i/2)
		}
		if prev != nil && string(prev) != string(cur) {
			changes++
		}
		prev = cur
	}
	clobber(8)
	if changes != 2 {
		t.Errorf("got %d changes, want 2", changes)
	}
	if string(prev) != "cccccccc" {
		t.Errorf("got %q, want cccccccc", prev)
	}
}

// Two makes in one function, one of which escapes.
//
//go:noinline
func twoMakes(n int) ([]byte, int) {
	local := make([]byte, size(n))
	kept := make([]byte, size(n))
	for i := range local {
		local[i] = 1
		kept[i] = 2
	}
	sum := 0
	for _, c := range local {
		sum += int(c)
	}
	return kept, sum
}

func TestTwoMakesInOneFunction(t *testing.T) {
	for _, n := range []int{4, 32, 40} {
		kept, sum := twoMakes(n)
		clobber(8)
		if sum != n {
			t.Errorf("length %d: sum %d, want %d", n, sum, n)
		}
		for _, c := range kept {
			if c != 2 {
				t.Fatalf("length %d: kept slice changed: %v", n, kept)
			}
		}
	}
}
