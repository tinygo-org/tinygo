//go:build !wasi && !wasip1 && !darwin

package syscall

func (e Errno) Is(target error) bool { return false }
