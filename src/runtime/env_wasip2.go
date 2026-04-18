//go:build wasip2

package runtime

// Notify the runtime when environment variables change.
// On wasip2, the environment is managed in Go (no C setenv), but
// internal/godebug still needs to be notified of GODEBUG changes.

//go:linkname syscallSetenv syscall.runtimeSetenv
func syscallSetenv(key, value string) {
	if key == "GODEBUG" && godebugUpdate != nil {
		godebugUpdate(key, value)
	}
}

//go:linkname syscallUnsetenv syscall.runtimeUnsetenv
func syscallUnsetenv(key string) {
	if key == "GODEBUG" && godebugUpdate != nil {
		godebugUpdate(key, "")
	}
}
