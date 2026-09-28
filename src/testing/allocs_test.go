package testing_test

import "testing"

var (
	allocations [101]*byte
	allocation  int
)

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
