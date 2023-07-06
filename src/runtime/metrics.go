package runtime

// Implementation of functions needed by runtime/metrics.
// Mostly just dummy implementations: we don't currently use any of these
// metrics.

//go:linkname godebug_registerMetric internal/godebug.registerMetric
func godebug_registerMetric(name string, read func() uint64) {
	// Dummy function for compatibility with Go 1.21.
}
