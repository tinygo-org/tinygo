//go:build tinygo

package escapeallocs

import "testing"

// These tests try to break the escape analysis rules. Tests named
// Unsafe* build objects that must stay on the heap, then clobber the stack and
// check that the values are intact. Tests named Safe* check that objects that
// can live on the stack do.

var (
	globalHalves [2]*uint64
	globalArr    [2]*[16]byte
	globalRunes  []rune
	globalBytes  []byte
	sinkSum      int
)

// clobber overwrites a large part of the stack below the caller.
//
//go:noinline
func clobber(n int) int {
	var pad [256]byte
	for i := range pad {
		pad[i] = 0xAA
	}
	if n > 0 {
		return clobber(n-1) + int(pad[n&127])
	}
	return int(pad[0])
}

func mustAllocate(t *testing.T, want float64, fn func()) {
	t.Helper()
	if allocs := testing.AllocsPerRun(20, fn); allocs < want {
		t.Errorf("got %v allocations, want at least %v", allocs, want)
	}
}

// Loops

// Nested loops that merge an object allocated outside the loops.
func TestSafeNestedLoopMerge(t *testing.T) {
	checkNoAllocs(t, func() {
		buf := make([]byte, 16)
		buf[3] = 7
		sum := 0
		for i := 0; i < 4; i++ {
			for j := 0; j < 4; j++ {
				p := buf
				if (i+j)&1 == 0 {
					p = other
				}
				sum += int(p[3])
			}
		}
		if sum != 8*7 {
			panic("wrong sum")
		}
	})
}

// An object allocated in an inner loop and kept for the next iteration.
func TestUnsafeAllocInInnerLoopKept(t *testing.T) {
	mustAllocate(t, 9, func() {
		sum := 0
		for i := 0; i < 3; i++ {
			var prev *[16]byte
			for j := 0; j < 3; j++ {
				b := new([16]byte)
				b[0] = byte(j + 1)
				if prev != nil {
					sum += int(prev[0])
				}
				prev = b
			}
		}
		if sum != 3*(1+2) {
			panic("wrong sum")
		}
	})
}

// A loop built with goto instead of for.
func TestUnsafeGotoLoopKept(t *testing.T) {
	mustAllocate(t, 5, func() {
		var prev *[16]byte
		sum := 0
		i := 0
	again:
		b := new([16]byte)
		b[0] = byte(i + 1)
		if prev != nil {
			sum += int(prev[0])
		}
		prev = b
		i++
		if i < 5 {
			goto again
		}
		if sum != 1+2+3+4 {
			panic("wrong sum")
		}
	})
}

// An object that is allocated in a loop and swapped with a second object.
func TestUnsafeAllocInLoopSwapped(t *testing.T) {
	mustAllocate(t, 6, func() {
		a, b := new([16]byte), new([16]byte)
		a[0], b[0] = 1, 2
		for i := 0; i < 6; i++ {
			c := new([16]byte)
			c[0] = byte(10 + i)
			a, b = b, a
			b = c
		}
		if a[0] == 0 || b[0] != 15 {
			panic("wrong value")
		}
	})
}

//go:noinline
func allocAndRead(i int) int {
	b := new([16]byte)
	b[0] = byte(i)
	return int(b[0])
}

// The callee allocates once per call, so it can use the stack even when it is
// called in a loop.
func TestSafeCalleeAllocCalledInLoop(t *testing.T) {
	checkNoAllocs(t, func() {
		sum := 0
		for i := 0; i < 8; i++ {
			sum += allocAndRead(i)
		}
		sinkSum = sum
	})
}

//go:noinline
func recurse(n int, prev *[16]byte) int {
	b := new([16]byte)
	b[0] = byte(n + 1)
	if n == 0 {
		return int(prev[0])
	}
	return recurse(n-1, b) + int(prev[0])
}

// Recursion that passes each frame's object to the next frame.
func TestRecursionPassesObject(t *testing.T) {
	first := new([16]byte)
	first[0] = 100
	if got, want := recurse(5, first), 6+5+4+3+2+100; got != want {
		t.Errorf("got %d, want %d", got, want)
	}
}

var globalRecursivePick *[16]byte

// pickRecursive gets the object as b, and the recursive call gets it as a.
// Both reach the same PHI, and only the recursive call's result is stored.
//
//go:noinline
func pickRecursive(a, b *[16]byte, n int) *[16]byte {
	var x *[16]byte
	if n&1 == 0 {
		sinkSum++
		x = a
	} else {
		sinkSum--
		x = b
	}
	if n > 0 {
		globalRecursivePick = pickRecursive(x, nil, n-1)
	}
	return x
}

//go:noinline
func passToPickRecursive() {
	p := new([16]byte)
	p[0] = 42
	pickRecursive(nil, p, 1)
}

// A recursive call reaches a PHI that is still being checked, through another
// parameter, and its result is stored in a global.
func TestUnsafeRecursivePhiResultStored(t *testing.T) {
	mustAllocate(t, 1, passToPickRecursive)
	passToPickRecursive()
	clobber(8)
	if globalRecursivePick[0] != 42 {
		t.Errorf("got %d, want 42", globalRecursivePick[0])
	}
}

// Callee loops

//go:noinline
func appendRunes(buf []rune, n int) []rune {
	for i := 0; i < n; i++ {
		buf = append(buf, rune(i+1))
	}
	return buf
}

func checkRunes(t *testing.T, got []rune, n int) {
	t.Helper()
	if len(got) != n {
		t.Fatalf("got length %d, want %d", len(got), n)
	}
	for i, r := range got {
		if r != rune(i+1) {
			t.Fatalf("element %d is %d, want %d", i, r, i+1)
		}
	}
}

// The appended buffer grows past its capacity, so the result is a new array
// and the original buffer is only read.
func TestAppendGrowsInCallee(t *testing.T) {
	mustAllocate(t, 1, func() {
		buf := make([]rune, 0, 4)
		out := appendRunes(buf, 20)
		checkRunes(t, out, 20)
	})
}

// The callee stores the buffer in a global and later calls must not change
// earlier results.
func TestUnsafeCalleeStoresBuffer(t *testing.T) {
	var kept [8][]rune
	mustAllocate(t, 1, func() {
		for i := range kept {
			buf := make([]rune, 0, 16)
			globalRunes = appendRunes(buf, i+1)
			kept[i] = globalRunes
		}
		clobber(8)
		for i := range kept {
			checkRunes(t, kept[i], i+1)
		}
	})
}

// The result of appending in a callee is kept across loop iterations.
func TestUnsafeReturnedBufferKeptInLoop(t *testing.T) {
	mustAllocate(t, 8, func() {
		var out [8][]rune
		for i := range out {
			buf := make([]rune, 0, 16)
			out[i] = appendRunes(buf, i+1)
		}
		clobber(8)
		for i := range out {
			checkRunes(t, out[i], i+1)
		}
	})
}

// The callee returns the buffer it was given and the caller carries it in a
// loop with other buffers.
func TestUnsafeBufferThroughPhiAndCall(t *testing.T) {
	mustAllocate(t, 6, func() {
		var cur []rune
		var all [6][]rune
		for i := range all {
			next := make([]rune, 0, 16)
			next = appendRunes(next, i+1)
			if cur != nil && len(cur) > len(next) {
				next = cur
			}
			cur = next
			all[i] = cur
		}
		clobber(8)
		for i := range all {
			checkRunes(t, all[i], i+1)
		}
	})
}

var (
	globalPicked *[16]byte
	otherArray   = new([16]byte)
)

// pickNested returns a, or b if any entry is set. a and b go through two PHIs
// that form a cycle, and b only reaches the return through the inner one.
//
//go:noinline
func pickNested(a, b *[16]byte, rows [][]bool) *[16]byte {
	p := a
	for _, row := range rows {
		q := p
		for _, set := range row {
			if set {
				q = b
			}
		}
		p = q
	}
	return p
}

//go:noinline
func keepPickedSecond(rows [][]bool) int {
	b := new([16]byte)
	b[0] = 42
	kept := pickNested(otherArray, b, rows)
	read := pickNested(b, otherArray, rows)
	globalPicked = kept
	return int(read[1])
}

//go:noinline
func keepPickedFirst(rows [][]bool) int {
	b := new([16]byte)
	b[0] = 42
	read := pickNested(b, otherArray, rows)
	kept := pickNested(otherArray, b, rows)
	globalPicked = kept
	return int(read[1])
}

// The same callee returns the object through two parameters. Whichever call is
// checked first must not hide that the other one returns it too.
func TestUnsafeReturnedThroughNestedLoops(t *testing.T) {
	rows := [][]bool{{false}, {false, true}}
	for _, keep := range []func([][]bool) int{keepPickedSecond, keepPickedFirst} {
		mustAllocate(t, 1, func() { keep(rows) })
		keep(rows)
		clobber(8)
		if globalPicked[0] != 42 {
			t.Errorf("got %d, want 42", globalPicked[0])
		}
	}
}

// Stack slots

type wide struct{ hi, lo uint64 }

func (u *wide) parts() [2]*uint64 {
	return [2]*uint64{&u.hi, &u.lo}
}

//go:noinline
func keepHalves(p *[2]*uint64) {
	globalHalves = *p
}

//go:noinline
func storeHalvesInGlobal(hi, lo uint64) {
	a := new(wide)
	a.hi, a.lo = hi, lo
	arr := [2]*uint64{&a.hi, &a.lo}
	keepHalves(&arr)
}

// The address of the array of pointers is passed to a function that copies
// the pointers into a global.
func TestUnsafePointerArrayAddressPassed(t *testing.T) {
	storeHalvesInGlobal(11, 22)
	clobber(8)
	if *globalHalves[0] != 11 || *globalHalves[1] != 22 {
		t.Errorf("got %d %d, want 11 22", *globalHalves[0], *globalHalves[1])
	}
}

//go:noinline
func returnHalves(hi, lo uint64) [2]*uint64 {
	a := new(wide)
	a.hi, a.lo = hi, lo
	arr := [2]*uint64{&a.hi, &a.lo}
	return arr
}

// The array of pointers into the object is returned.
func TestUnsafePointerArrayReturned(t *testing.T) {
	h := returnHalves(33, 44)
	clobber(8)
	if *h[0] != 33 || *h[1] != 44 {
		t.Errorf("got %d %d, want 33 44", *h[0], *h[1])
	}
}

//go:noinline
func copyHalvesToGlobal(hi, lo uint64) {
	a := new(wide)
	a.hi, a.lo = hi, lo
	globalHalves = [2]*uint64{&a.hi, &a.lo}
}

// The array of pointers is stored directly in a global.
func TestUnsafePointerArrayStoredInGlobal(t *testing.T) {
	copyHalvesToGlobal(55, 66)
	clobber(8)
	if *globalHalves[0] != 55 || *globalHalves[1] != 66 {
		t.Errorf("got %d %d, want 55 66", *globalHalves[0], *globalHalves[1])
	}
}

//go:noinline
func pickDynamic(i int) {
	a, b := new([16]byte), new([16]byte)
	a[0], b[0] = byte(i), byte(i+1)
	arr := [2]*[16]byte{a, b}
	globalArr[i&1] = arr[(i+1)&1]
}

// A pointer picked from a local array with a run time index is stored in a
// global.
func TestUnsafePointerPickedByIndexStored(t *testing.T) {
	pickDynamic(4)
	pickDynamic(7)
	clobber(8)
	if globalArr[0][0] != 5 || globalArr[1][0] != 7 {
		t.Errorf("got %d %d, want 5 7", globalArr[0][0], globalArr[1][0])
	}
}

//go:noinline
func slotInClosure(i int) func() int {
	a := new([16]byte)
	a[0] = byte(i)
	arr := [2]*[16]byte{a, a}
	return func() int { return int(arr[i&1][0]) }
}

// The array of pointers is captured by a closure that outlives the function.
func TestUnsafePointerArrayCapturedByClosure(t *testing.T) {
	f := slotInClosure(9)
	clobber(8)
	if got := f(); got != 9 {
		t.Errorf("got %d, want 9", got)
	}
}

type pair struct{ a, b *[16]byte }

//go:noinline
func structSlotEscapes(i int) *pair {
	x, y := new([16]byte), new([16]byte)
	x[0], y[0] = byte(i), byte(i+1)
	p := pair{x, y}
	return &p
}

// The pointers are held in a struct that escapes.
func TestUnsafePointerInEscapingStruct(t *testing.T) {
	p := structSlotEscapes(3)
	clobber(8)
	if p.a[0] != 3 || p.b[0] != 4 {
		t.Errorf("got %d %d, want 3 4", p.a[0], p.b[0])
	}
}

// A pointer picked with a run time index from a local array that never
// leaves the function.
func TestSafePointerPickedByIndex(t *testing.T) {
	checkNoAllocs(t, func() {
		var w, x, y, z [16]byte
		w[0], x[0], y[0], z[0] = 1, 2, 3, 4
		arr := [4]*[16]byte{new([16]byte), new([16]byte), new([16]byte), new([16]byte)}
		arr[0][0], arr[1][0], arr[2][0], arr[3][0] = w[0], x[0], y[0], z[0]
		sum := 0
		for i := 0; i < 8; i++ {
			sum += int(arr[i&3][0])
		}
		if sum != 2*(1+2+3+4) {
			panic("wrong sum")
		}
	})
}

// Strings to bytes

const constString = "abcdefghijklmnopqrstuvwxyz0123456789ABCDEF"

func checkConstString(t *testing.T) {
	t.Helper()
	if got := string([]byte(constString)); got != constString {
		t.Fatalf("constant string was modified: %q", got)
	}
}

//go:noinline
func fillByReslicing(b []byte) {
	for len(b) > 0 {
		b[0] = 'x'
		b = b[1:]
	}
}

// A callee that writes through a reslicing loop must get a copy.
func TestUnsafeWriteThroughReslice(t *testing.T) {
	mustAllocate(t, 1, func() { fillByReslicing([]byte(constString)) })
	fillByReslicing([]byte(constString))
	checkConstString(t)
}

//go:noinline
func writeLast(b []byte) { b[len(b)-1] = '!' }

//go:noinline
func chain1(b []byte) { writeLast(b) }

//go:noinline
func chain2(b []byte) { chain1(b) }

//go:noinline
func chain3(b []byte) { chain2(b) }

//go:noinline
func chain4(b []byte) { chain3(b) }

//go:noinline
func chain5(b []byte) { chain4(b) }

//go:noinline
func chain6(b []byte) { chain5(b) }

// The write is deeper in the call chain than the analysis looks.
func TestUnsafeWriteDeepInCallChain(t *testing.T) {
	chain1([]byte(constString))
	chain2([]byte(constString))
	chain3([]byte(constString))
	chain4([]byte(constString))
	chain5([]byte(constString))
	chain6([]byte(constString))
	checkConstString(t)
}

//go:noinline
func readThenWriteLater(b []byte, write bool) int {
	sum := 0
	for ; len(b) > 8; b = b[8:] {
		sum += int(b[0])
	}
	if write {
		b[0] = '?'
	}
	return sum
}

// The write is on a path that is not taken in the first call.
func TestUnsafeWriteOnOtherPath(t *testing.T) {
	readThenWriteLater([]byte(constString), false)
	readThenWriteLater([]byte(constString), true)
	checkConstString(t)
}

//go:noinline
func keepSlice(b []byte) int {
	n := 0
	for ; len(b) > 8; b = b[8:] {
		n++
	}
	globalBytes = b
	return n
}

// The callee keeps the slice and the caller writes to it later.
func TestUnsafeCalleeKeepsSlice(t *testing.T) {
	keepSlice([]byte(constString))
	globalBytes[0] = 'Z'
	checkConstString(t)
}

//go:noinline
func returnSuffix(b []byte) []byte {
	for len(b) > 4 {
		b = b[1:]
	}
	return b
}

// The callee returns part of the slice and the caller writes to it.
func TestUnsafeCalleeReturnsSlice(t *testing.T) {
	r := returnSuffix([]byte(constString))
	r[0] = 'Q'
	checkConstString(t)
}

//go:noinline
func passToWriter(b []byte, w func([]byte)) int {
	for ; len(b) > 8; b = b[8:] {
	}
	w(b)
	return len(b)
}

// The callee passes the slice to an indirect call that writes.
func TestUnsafeCalleeCallsIndirect(t *testing.T) {
	passToWriter([]byte(constString), func(b []byte) { b[0] = '#' })
	checkConstString(t)
}

//go:noinline
func readRecursive(b []byte) int {
	if len(b) <= 8 {
		return len(b)
	}
	return readRecursive(b[8:]) + int(b[0])
}

//go:noinline
func writeRecursive(b []byte) int {
	if len(b) <= 8 {
		b[0] = '%'
		return len(b)
	}
	return writeRecursive(b[8:]) + int(b[0])
}

// Recursive callees, one reads and one writes.
func TestRecursiveCallees(t *testing.T) {
	readRecursive([]byte(constString))
	writeRecursive([]byte(constString))
	checkConstString(t)
}

//go:noinline
func readOnlyChain(b []byte) int { return readOnlyLeaf(b) + len(b) }

//go:noinline
func readOnlyLeaf(b []byte) int {
	n := 0
	for ; len(b) >= 4; b = b[4:] {
		n += int(b[0])
	}
	return n
}

// Callees that only read, through a loop and through a call, allocate nothing.
func TestSafeReadOnlyChain(t *testing.T) {
	checkNoAllocs(t, func() {
		sinkSum = readOnlyChain([]byte(constString)) + readOnlyLeaf([]byte(constString))
	})
}

// Compile time

var mergeConds [32]bool

//go:noinline
func stepSlice(b []byte) []byte { return b[1:] }

// manyMergesBeforeLoop has a chain of merge PHIs before a loop. The analysis
// reaches each merge PHI on two paths, so it takes exponential time unless
// the results for the merge PHIs are reused.
//
//go:noinline
func manyMergesBeforeLoop() int {
	buf := make([]byte, 64)
	for i := range buf {
		buf[i] = 1
	}
	b := buf
	if mergeConds[0] {
		b = stepSlice(b)
	}
	if mergeConds[1] {
		b = stepSlice(b)
	}
	if mergeConds[2] {
		b = stepSlice(b)
	}
	if mergeConds[3] {
		b = stepSlice(b)
	}
	if mergeConds[4] {
		b = stepSlice(b)
	}
	if mergeConds[5] {
		b = stepSlice(b)
	}
	if mergeConds[6] {
		b = stepSlice(b)
	}
	if mergeConds[7] {
		b = stepSlice(b)
	}
	if mergeConds[8] {
		b = stepSlice(b)
	}
	if mergeConds[9] {
		b = stepSlice(b)
	}
	if mergeConds[10] {
		b = stepSlice(b)
	}
	if mergeConds[11] {
		b = stepSlice(b)
	}
	if mergeConds[12] {
		b = stepSlice(b)
	}
	if mergeConds[13] {
		b = stepSlice(b)
	}
	if mergeConds[14] {
		b = stepSlice(b)
	}
	if mergeConds[15] {
		b = stepSlice(b)
	}
	if mergeConds[16] {
		b = stepSlice(b)
	}
	if mergeConds[17] {
		b = stepSlice(b)
	}
	if mergeConds[18] {
		b = stepSlice(b)
	}
	if mergeConds[19] {
		b = stepSlice(b)
	}
	if mergeConds[20] {
		b = stepSlice(b)
	}
	if mergeConds[21] {
		b = stepSlice(b)
	}
	if mergeConds[22] {
		b = stepSlice(b)
	}
	if mergeConds[23] {
		b = stepSlice(b)
	}
	if mergeConds[24] {
		b = stepSlice(b)
	}
	if mergeConds[25] {
		b = stepSlice(b)
	}
	if mergeConds[26] {
		b = stepSlice(b)
	}
	if mergeConds[27] {
		b = stepSlice(b)
	}
	if mergeConds[28] {
		b = stepSlice(b)
	}
	if mergeConds[29] {
		b = stepSlice(b)
	}
	if mergeConds[30] {
		b = stepSlice(b)
	}
	if mergeConds[31] {
		b = stepSlice(b)
	}
	s := 0
	for len(b) > 0 {
		s += int(b[0])
		b = b[1:]
	}
	return s
}

func TestManyMergesBeforeLoop(t *testing.T) {
	for i := range mergeConds {
		mergeConds[i] = i%2 == 0
	}
	if got := manyMergesBeforeLoop(); got != 64-16 {
		t.Errorf("got %d, want %d", got, 64-16)
	}
	checkNoAllocs(t, func() { sinkSum = manyMergesBeforeLoop() })
}
