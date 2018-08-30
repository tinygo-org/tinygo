package runtime

// This file contains support code for the sync package.

//go:linkname registerPoolCleanup sync.runtime_registerPoolCleanup
func registerPoolCleanup(cleanup func()) {
	// Ignore.
}

//go:linkname notifyListCheck sync.runtime_notifyListCheck
func notifyListCheck(size uintptr) {
	// Ignore.
}
