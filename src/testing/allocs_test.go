package testing_test

import (
	"runtime"
	"testing"
)

var (
	allocations [101]*byte
	allocation  int

	memStatsAllocations [101]*byte
	memStatsAllocation  int
)

func TestMemStatsMallocs(t *testing.T) {
	var before, after runtime.MemStats
	runtime.ReadMemStats(&before)
	allocs := testing.AllocsPerRun(100, func() {
		memStatsAllocations[memStatsAllocation] = new(byte)
		memStatsAllocation++
	})
	runtime.ReadMemStats(&after)
	if got, want := after.Mallocs-before.Mallocs, uint64(allocs)*101; got != want {
		t.Errorf("got %d mallocs, want %d", got, want)
	}
	for i := 1; i < len(memStatsAllocations); i++ {
		if memStatsAllocations[i] == memStatsAllocations[i-1] {
			t.Fatalf("allocations %d and %d have the same address", i-1, i)
		}
	}
}

func TestAllocsPerRun(t *testing.T) {
	allocs := testing.AllocsPerRun(100, func() {
		allocations[allocation] = new(byte)
		allocation++
	})
	if allocs != 1 {
		t.Errorf("got %v allocations, want 1", allocs)
	}
	for i := 1; i < len(allocations); i++ {
		if allocations[i] == allocations[i-1] {
			t.Fatalf("allocations %d and %d have the same address", i-1, i)
		}
	}
}
