//go:build tinygo

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

//go:linkname syscall_fcntl runtime/runtime.fcntl
func syscall_fcntl(fd, cmd, arg int32) (ret int32, errno int32) {
	// https://cs.opensource.google/go/go/+/master:src/runtime/os_linux.go;l=452?q=runtime.fcntl&ss=go%2Fgo
	r, _, err := Syscall6(SYS_FCNTL, uintptr(fd), uintptr(cmd), uintptr(arg), 0, 0, 0)
	return int32(r), int32(err)
}
