//go:build tinygo

package syscall

var system systemer = &noSystem{}

func useSystem(s systemer) {
	system = s
}

type systemer interface {
	Socket(domain, typ, proto int) (fd int, err error)
	CloseOnExec(fd int)
	SetNonblock(fd int, nonblocking bool) (err error)
	SetsockoptInt(fd, level, opt int, value int) (err error)
	Connect(fd int, ip []byte, port uint16) (err error)
	Write(fd int, buf []byte) (n int, err error)
	Read(fd int, buf []byte) (n int, err error)
	Getsockname(fd int) (ip []byte, port uint16, err error)
	Getpeername(fd int) (ip []byte, port uint16, err error)
	Close(fd int) error
}

type noSystem struct{}

func (s noSystem) Socket(domain, typ, proto int) (fd int, err error) {
	println("Socket not implemented", domain, typ, proto)
	return fd, ENOSYS
}

func (s noSystem) CloseOnExec(fd int)                                      {}
func (s noSystem) SetNonblock(fd int, nonblocking bool) (err error)        { return ENOSYS }
func (s noSystem) SetsockoptInt(fd, level, opt int, value int) (err error) { return ENOSYS }
func (s noSystem) Connect(fd int, ip []byte, port uint16) (err error)      { return ENOSYS }
func (s noSystem) Write(fd int, buf []byte) (n int, err error)             { return n, ENOSYS }
func (s noSystem) Read(fd int, buf []byte) (n int, err error)              { return n, ENOSYS }
func (s noSystem) Getsockname(fd int) (ip []byte, port uint16, err error)  { return ip, port, ENOSYS }
func (s noSystem) Getpeername(fd int) (ip []byte, port uint16, err error)  { return ip, port, ENOSYS }
func (s noSystem) Close(fd int) error                                      { return ENOSYS }
