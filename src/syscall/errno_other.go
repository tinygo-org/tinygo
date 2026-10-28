//go:build !js && !wasip1 && !wasip2 && !windows

package syscall

func (e Errno) Is(target error) bool { return false }
