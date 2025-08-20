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
