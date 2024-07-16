package syscall

import (
	"errors"
)

const (
	MSG_DONTWAIT = 0x40
	AF_INET      = 0x2
	AF_INET6     = 0xa
)

func Exit(code int)

type Rlimit struct {
	Cur uint64
	Max uint64
}

func Setrlimit(resource int, rlim *Rlimit) error {
	return errors.New("Setrlimit not implemented")
}
