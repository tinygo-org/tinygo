//go:build tinygo

package main

import (
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/tinygo-org/tinygo/tools/qbs/dev"
)

var sink any

// noReclaim is set when the collector does not give memory back.
var noReclaim bool

// extraTests holds tests that need more flash or RAM than every target has.
var extraTests []dev.Test

func main() {
	tests := []dev.Test{
		{Name: "reclaims", Run: reclaims},
		{Name: "large-block", Run: largeBlock},
		{Name: "linked-list", Run: linkedList},
		{Name: "slice-grow", Run: sliceGrow},
		{Name: "strings", Run: stringsTest},
		{Name: "gc-cycles", Run: gcCycles},
	}
	dev.Main("gc", append(tests, extraTests...))
}

func stats() runtime.MemStats {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	return m
}

// clobberStack overwrites the stack and registers so that stale pointers from
// earlier calls do not keep memory alive in a conservative collector.
//
//go:noinline
func clobberStack(n int) byte {
	var buf [64]byte
	for i := range buf {
		buf[i] = byte(i + n)
	}
	if n > 0 {
		return buf[n%64] + clobberStack(n-1)
	}
	return buf[0]
}

// churn allocates a few small objects. A conservative collector can keep the
// most recent large allocation alive through a stale pointer until later
// allocations overwrite it.
//
//go:noinline
func churn() {
	for i := 0; i < 8; i++ {
		sink = make([]byte, 16+i*3)
	}
	sink = nil
}

func collect() {
	sink = clobberStack(4)
	sink = nil
	churn()
	runtime.GC()
}

// maxHeap limits the heap the tests use. A target with a very large heap, such
// as an emulator, would otherwise make them slow, and a conservative collector
// retains more of a large heap by mistaking data for pointers.
const maxHeap = 4 << 20

func freeHeap() int {
	collect()
	m := stats()
	if m.HeapIdle > maxHeap {
		return maxHeap
	}
	return int(m.HeapIdle)
}

func inUse() int {
	collect()
	m := stats()
	return int(m.HeapInuse)
}

// budget returns the number of bytes a test may allocate, or 0 when the test
// should not run.
func budget(t *dev.T, divisor int) int {
	if noReclaim {
		t.Skip("the collector does not reclaim memory")
		return 0
	}
	b := freeHeap() / divisor
	if b < 512 {
		t.Skip("heap too small")
		return 0
	}
	return b
}

// checkLeak fails the test when much more than the largest object is still in
// use. A conservative collector can keep the most recent large object alive
// through a stale pointer.
func checkLeak(t *dev.T, before, allocated, largest int) {
	after := inUse()
	t.Val("retained_bytes", strconv.Itoa(after-before))
	if after-before > largest+allocated/4+1024 {
		t.Fail("memory not released, " + strconv.Itoa(after-before) + " of " + strconv.Itoa(allocated) + " bytes retained")
	}
}

// noGrowth runs f five times and fails if memory use grows on each of the last
// three runs. A stale pointer can keep a whole structure alive, and which
// pointers are stale differs from run to run, so memory use may jump once or
// twice. A real leak adds a structure on every run.
func noGrowth(t *dev.T, what string, size int, f func() bool) {
	var used [5]int
	for i := range used {
		if !f() {
			t.Fail(what + " was corrupted")
			return
		}
		used[i] = inUse()
		t.Val("in_use_after_run_"+strconv.Itoa(i+1), strconv.Itoa(used[i]))
	}
	step := size / 4
	if used[2]-used[1] > step && used[3]-used[2] > step && used[4]-used[3] > step {
		t.Fail("memory use grew on every run, from " + strconv.Itoa(used[1]) + " to " + strconv.Itoa(used[4]) + " bytes")
	}
}

//go:noinline
func allocAndDrop(n int) {
	b := make([]byte, n)
	for i := range b {
		b[i] = byte(i)
	}
	sink = b
	sink = nil
}

func reclaims(t *dev.T) {
	free := freeHeap()
	t.Val("heap_free_bytes", strconv.Itoa(free))
	t.Val("heap_in_use_bytes", strconv.Itoa(inUse()))
	n := free / 8
	if n < 512 {
		noReclaim = true
		t.Skip("heap too small")
		return
	}
	// A collector that frees memory holds at most the last block afterwards.
	before := inUse()
	for i := 0; i < 3; i++ {
		allocAndDrop(n + i*16)
	}
	retained := inUse() - before
	t.Val("retained_bytes", strconv.Itoa(retained))
	if retained > 2*n {
		noReclaim = true
		t.Skip("the collector does not reclaim memory")
	}
}

//go:noinline
func bigBlock(n int) bool {
	b := make([]byte, n)
	for i := range b {
		b[i] = byte(i * 13)
	}
	for i := range b {
		if b[i] != byte(i*13) {
			return false
		}
	}
	return true
}

func largeBlock(t *dev.T) {
	n := budget(t, 3)
	if n == 0 {
		return
	}
	before := inUse()
	if !bigBlock(n) {
		t.Fail("contents of a " + strconv.Itoa(n) + " byte block changed")
	}
	t.Val("retained_bytes", strconv.Itoa(inUse()-before))
}

type node struct {
	next *node
	id   uint32
	pad  [4]uint32
}

//go:noinline
func listTest(n int) bool {
	var head *node
	for i := 0; i < n; i++ {
		head = &node{next: head, id: uint32(i) * 2654435761}
	}
	for i := n - 1; i >= 0; i-- {
		if head == nil || head.id != uint32(i)*2654435761 {
			return false
		}
		head = head.next
	}
	return head == nil
}

func linkedList(t *dev.T) {
	b := budget(t, 4)
	if b == 0 {
		return
	}
	n := b / 40
	noGrowth(t, "linked list", n*32, func() bool { return listTest(n) })
}

//go:noinline
func growTest(limit int) bool {
	var s []byte
	for i := 0; i < limit; i++ {
		s = append(s, byte(i*7))
	}
	for i := range s {
		if s[i] != byte(i*7) {
			return false
		}
	}
	return len(s) == limit
}

func sliceGrow(t *dev.T) {
	b := budget(t, 4)
	if b == 0 {
		return
	}
	limit := b / 3
	noGrowth(t, "appended data", limit*2, func() bool { return growTest(limit) })
}

//go:noinline
func stringsCheck(count int) bool {
	var sb strings.Builder
	for i := 0; i < count; i++ {
		sb.WriteString(strconv.Itoa(i))
		sb.WriteByte(',')
	}
	s := sb.String()
	pos := 0
	for i := 0; i < count; i++ {
		d := strconv.Itoa(i) + ","
		if pos+len(d) > len(s) || s[pos:pos+len(d)] != d {
			return false
		}
		pos += len(d)
	}
	return pos == len(s)
}

// stringsTest builds a long decimal list.
func stringsTest(t *dev.T) {
	b := budget(t, 6)
	if b == 0 {
		return
	}
	count := b / 24
	before := inUse()
	if !stringsCheck(count) {
		t.Fail("string builder output was wrong")
	}
	checkLeak(t, before, b, b/2)
}

//go:noinline
func garbage(n int) {
	sink = make([]byte, n)
}

func gcCycles(t *dev.T) {
	b := budget(t, 8)
	if b == 0 {
		return
	}
	size := 64 + 64*(b/4096%8)
	if size > b/4 {
		size = b / 4
	}
	runtime.GC()
	start := stats()
	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		garbage(size)
		m := stats()
		if m.TotalAlloc-start.TotalAlloc > 3*m.HeapSys && m.NumGC > start.NumGC {
			break
		}
	}
	sink = nil
	end := stats()
	t.Val("collections", strconv.Itoa(int(end.NumGC-start.NumGC)))
	t.Val("allocated_bytes", strconv.Itoa(int(end.TotalAlloc-start.TotalAlloc)))
	if end.NumGC == start.NumGC {
		allocated := end.TotalAlloc - start.TotalAlloc
		if end.HeapSys > 8*allocated {
			t.Skip("the heap is too large to reach a collection in a few seconds")
			return
		}
		t.Fail("no collection ran while allocating " + strconv.Itoa(int(allocated)) + " bytes")
	}
}
