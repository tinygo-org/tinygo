//go:build linux && (baremetal || nintendoswitch || wasm_unknown)

// Other systems that aren't operating systems supported by the Go toolchain
// need to pretend to be an existing operating system. Linux seems like a good
// choice for this for its wide hardware support.

package runtime

const GOOS = "linux"
