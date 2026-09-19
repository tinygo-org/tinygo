package main

import "syscall"

func main() {
	fd, err := syscall.Socket(syscall.AF_INET, syscall.SOCK_STREAM, 0)
	check(err)
	defer syscall.Close(fd)
	check(syscall.SetsockoptInt(fd, syscall.SOL_SOCKET, syscall.SO_REUSEADDR, 1))
	check(syscall.Bind(fd, &syscall.SockaddrInet4{Addr: [4]byte{127, 0, 0, 1}}))
	check(syscall.Listen(fd, 1))
	addr, err := syscall.Getsockname(fd)
	check(err)
	if addr.(*syscall.SockaddrInet4).Port == 0 {
		panic("the listener has no port")
	}
	client, err := syscall.Socket(syscall.AF_INET, syscall.SOCK_STREAM, 0)
	check(err)
	defer syscall.Close(client)
	check(syscall.Connect(client, addr))
	peer, _, err := syscall.Accept(fd)
	check(err)
	defer syscall.Close(peer)
	_, err = syscall.Getpeername(peer)
	check(err)
	check(syscall.Shutdown(client, syscall.SHUT_RDWR))
	println("Darwin socket symbols linked and ran")
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}
