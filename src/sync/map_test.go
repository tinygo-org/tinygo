package sync_test

import (
	"sync"
	"testing"
)

func TestMapLoadAndDelete(t *testing.T) {
	var sm sync.Map
	sm.Store("present", "value")

	if v, ok := sm.LoadAndDelete("present"); !ok || v != "value" {
		t.Errorf("LoadAndDelete returned %v, %v, want value, true", v, ok)
	}

	if v, ok := sm.LoadAndDelete("absent"); ok || v != nil {
		t.Errorf("LoadAndDelete returned %v, %v, want nil, false", v, ok)
	}
}

func TestMapSwap(t *testing.T) {
	var sm sync.Map
	sm.Store("present", "value")

	if v, ok := sm.Swap("present", "value2"); !ok || v != "value" {
		t.Errorf("Swap returned %v, %v, want value, true", v, ok)
	}
	if v, ok := sm.Load("present"); !ok || v != "value2" {
		t.Errorf("Load after Swap returned %v, %v, want value2, true", v, ok)
	}

	if v, ok := sm.Swap("new", "foo"); ok || v != nil {
		t.Errorf("Swap returned %v, %v, want nil, false", v, ok)
	}
	if v, ok := sm.Load("present"); !ok || v != "value2" {
		t.Errorf("Load after Swap returned %v, %v, want foo, true", v, ok)
	}
}

func TestMapRangeAndDelete(t *testing.T) {
	var sm sync.Map
	sm.Store(0, "0")
	sm.Store(1, "1")
	sm.Store(2, "2")

	sm.Range(func(k, v any) bool {
		keyAsInt, ok := k.(int)
		if !ok {
			return true
		}
		if keyAsInt%2 == 0 {
			sm.Delete(keyAsInt)
		}
		return true
	})

	if v, ok := sm.Load(0); ok {
		t.Errorf("Load(0) after Delete returned %v, %v, want nil, false", v, ok)
	}
	if v, ok := sm.Load(1); !ok || v.(string) != "1" {
		t.Errorf("Load(1) after Delete returned %v, %v, want \"1\", true", v, ok)
	}
	if v, ok := sm.Load(2); ok {
		t.Errorf("Load(2) after Delete returned %v, %v, want nil, false", v, ok)
	}
}
