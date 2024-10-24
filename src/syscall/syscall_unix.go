//go:build linux || unix
// +build linux unix

package syscall

func Exec(argv0 string, argv []string, envv []string) (err error)

// The two SockaddrInet* structs have been copied from the Go source tree.

type SockaddrInet4 struct {
	Port int
	Addr [4]byte
	raw  RawSockaddrInet4
}

type SockaddrInet6 struct {
	Port   int
	ZoneId uint32
	Addr   [16]byte
	raw    RawSockaddrInet6
}

const (
	SYS_FORK   = 57
	SYS_EXECVE = 59
)
