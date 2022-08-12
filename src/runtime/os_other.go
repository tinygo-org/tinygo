//go:build linux && (baremetal || nintendoswitch || wasi || unknow_wasm)
// +build linux
// +build baremetal nintendoswitch wasi unknow_wasm

// Other systems that aren't operating systems supported by the Go toolchain
// need to pretend to be an existing operating system. Linux seems like a good
// choice for this for its wide hardware support.

package runtime

const GOOS = "linux"
