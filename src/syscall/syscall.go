package syscall

import (
	"errors"

	"github.com/tinygo-org/tinygo/src/internal/bytealg"
)

const (
	MSG_DONTWAIT = 0x40
	AF_INET      = 0x2
	AF_INET6     = 0xa
)

const (
	// #define	EINVAL	22	/* Invalid argument			*/
	EINVAL = 22
)

func Exit(code int)

type Rlimit struct {
	Cur uint64
	Max uint64
}

func Setrlimit(resource int, rlim *Rlimit) error {
	return errors.New("Setrlimit not implemented")
}

// ByteSliceFromString returns a NUL-terminated slice of bytes
// containing the text of s. If s contains a NUL byte at any
// location, it returns (nil, [EINVAL]).
// https://cs.opensource.google/go/go/+/master:src/syscall/syscall.go;l=45;drc=94982a07825aec711f11c97283e99e467838d616
func ByteSliceFromString(s string) ([]byte, error) {
	if bytealg.IndexByteString(s, 0) != -1 {
		return nil, errors.New("contains NUL")
	}
	a := make([]byte, len(s)+1)
	copy(a, s)
	return a, nil
}

// BytePtrFromString returns a pointer to a NUL-terminated array of
// bytes containing the text of s. If s contains a NUL byte at any
// location, it returns (nil, [EINVAL]).
func BytePtrFromString(s string) (*byte, error) {
	a, err := ByteSliceFromString(s)
	if err != nil {
		return nil, err
	}
	return &a[0], nil
}
