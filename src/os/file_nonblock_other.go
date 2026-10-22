//go:build !linux || baremetal || wasm_unknown || nintendoswitch || wasip1 || wasip2 || tinygo.wasm

package os

// nbState is empty where Read and Write do not wait on non-blocking fds.
type nbState struct{}

func (f *file) wouldBlock(h FileHandle, err error) bool { return false }

func (f *file) waitIO(h FileHandle, write bool) error { return ErrNotImplemented }

func (f *file) closeDeferred() bool { return false }
