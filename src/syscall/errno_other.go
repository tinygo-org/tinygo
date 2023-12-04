//go:build !wasip1 && !wasip2 && !darwin

package syscall

func (e Errno) Is(target error) bool { return false }
