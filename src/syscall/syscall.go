package syscall

import (
	"errors"
	"sync/atomic"
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

// origRlimitNofile, if not {0, 0}, is the original soft RLIMIT_NOFILE.
// When we can assume that we are bootstrapping with Go 1.19,
// this can be atomic.Pointer[Rlimit].
var origRlimitNofile atomic.Value // of Rlimit

func Setrlimit(resource int, rlim *Rlimit) error {
	return errors.New("Setrlimit not implemented")
}
