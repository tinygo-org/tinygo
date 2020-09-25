package os

import (
	"syscall"
)

func Getenv(key string) string {
	v, _ := syscall.Getenv(key)
	return v
}

func LookupEnv(key string) (string, bool) {
	return syscall.Getenv(key)
}

func Environ() []string {
	return syscall.Environ()
}
