package runtime

// Dummy implementation of synctest functions (we don't support synctest at the
// moment).

//go:linkname synctest_acquire internal/synctest.acquire
func synctest_acquire() any {
	// Dummy: we don't support synctest.
	return nil
}

//go:linkname synctest_release internal/synctest.release
func synctest_release(sg any) {
	// Dummy: we don't support synctest.
}
