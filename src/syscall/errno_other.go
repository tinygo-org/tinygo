//go:build !wasip1 && !wasip2

package syscall

func (e Errno) Is(target error) bool { return false }
