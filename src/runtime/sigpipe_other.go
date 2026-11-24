//go:build !(darwin || (linux && !baremetal && !tinygo.wasm && !nintendoswitch))

package runtime

// Copied from the Go runtime source code.
//
//go:linkname os_sigpipe os.sigpipe
func os_sigpipe() {
	runtimeFatal("too many writes on closed pipe")
}
