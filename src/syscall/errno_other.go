//go:build !js && !tinygo.wasi && !windows

package syscall

func (e Errno) Is(target error) bool { return false }
