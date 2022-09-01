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
