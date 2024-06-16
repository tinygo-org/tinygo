//go:build tinygo

package syscall

func Listen(fd int, backlog int) error {
	println("syscall.Listen not implemented", fd, backlog)
	return ENOSYS
}

func Accept(fd int) (int, Sockaddr, error) {
	println("syscall.Accept not implemented", fd)
	return 0, nil, ENOSYS
}

func Shutdown(fd int, how int) error {
	println("syscall.Shutdown", fd, how)
	return ENOSYS
}
