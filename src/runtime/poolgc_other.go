//go:build !gc.boehm && !gc.conservative && !gc.precise

package runtime

// poolGCCount is always zero here, so sync.Pool keeps its items.
func poolGCCount() uint32 { return 0 }
