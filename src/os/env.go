package os

import (
	"syscall"
)

func Getenv(key string) string {
	v, _ := syscall.Getenv(key)
	return v
}

func Setenv(key, value string) error {
	err := syscall.Setenv(key, value)
	if err != nil {
		return NewSyscallError("setenv", err)
	}
	return nil
}

func Unsetenv(key string) error {
	err := syscall.Unsetenv(key)
	if err != nil {
		return NewSyscallError("unsetenv", err)
	}
	return nil
}

func LookupEnv(key string) (string, bool) {
	return syscall.Getenv(key)
}

func Environ() []string {
	return syscall.Environ()
}
