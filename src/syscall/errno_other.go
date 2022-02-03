//go:build !wasi && !darwin
// +build !wasi,!darwin

package syscall

func (e Errno) Is(target error) bool { return false }
