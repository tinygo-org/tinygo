package sync

// Stubs for upstream runtime-internal sync test helpers. Tests that rely on
// these are skipped in tools/stdlibtest/rules.go.
func Runtime_Semacquire(*uint32) {
	panic("sync.Runtime_Semacquire is gc-specific")
}

func Runtime_Semrelease(*uint32, bool, int) {
	panic("sync.Runtime_Semrelease is gc-specific")
}

func Runtime_procPin() int {
	panic("sync.Runtime_procPin is gc-specific")
}

func Runtime_procUnpin() {
	panic("sync.Runtime_procUnpin is gc-specific")
}

type PoolDequeue interface {
	PushHead(val any) bool
	PopHead() (any, bool)
	PopTail() (any, bool)
}

func NewPoolDequeue(n int) PoolDequeue {
	panic("sync.NewPoolDequeue is gc-specific")
}

func NewPoolChain() PoolDequeue {
	panic("sync.NewPoolChain is gc-specific")
}
