package sync

// Stubs for upstream runtime-internal sync test helpers. Tests that rely on
// these are skipped in tools/stdlibtest/rules.go.
func Runtime_Semacquire(*uint32) {
	panic("sync.Runtime_Semacquire is gc-specific")
}

func Runtime_Semrelease(*uint32, bool, int) {
	panic("sync.Runtime_Semrelease is gc-specific")
}
