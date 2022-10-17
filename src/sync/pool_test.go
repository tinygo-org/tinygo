package sync_test

import (
	"sync"
	"testing"
)

type testItem struct {
	val int
}

func TestPool(t *testing.T) {
	p := sync.Pool{
		New: func() interface{} {
			return &testItem{}
		},
	}

	i1P := p.Get()
	if i1P == nil {
		t.Error("pool with New returned nil")
	}
	i1 := i1P.(*testItem)
	if got, want := i1.val, 0; got != want {
		t.Errorf("empty pool item value: got %v, want %v", got, want)
	}
	i1.val = 1

	i2 := p.Get().(*testItem)
	if got, want := i2.val, 0; got != want {
		t.Errorf("empty pool item value: got %v, want %v", got, want)
	}
	i2.val = 2

	p.Put(i1)

	i3 := p.Get().(*testItem)
	if got, want := i3.val, 1; got != want {
		t.Errorf("pool with item value: got %v, want %v", got, want)
	}
}

func TestPool_noNew(t *testing.T) {
	p := sync.Pool{}

	i1 := p.Get()
	if i1 != nil {
		t.Errorf("pool without New returned %v, want nil", i1)
	}
}
