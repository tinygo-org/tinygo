//go:build !(darwin || (linux && !baremetal && !wasip1 && !wasm_unknown && !wasip2 && !nintendoswitch))

package runtime

// Copied from the Go runtime source code.
//
//go:linkname os_sigpipe os.sigpipe
func os_sigpipe() {
	runtimeFatal("too many writes on closed pipe")
}
