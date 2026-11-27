//go:build !(gc.conservative || gc.precise)

package runtime

// Memory profiling is not supported with this collector, so the profile is
// always empty.
func memProfile(p []MemProfileRecord, inuseZero bool) (n int, ok bool) {
	return 0, true
}
