package main

import (
	"io"
	"net"
	"strconv"
)

// Test golang network package integration for tinygo.
// This test is not exhaustive and only tests the basic functionality of the package.

var (
	testsPassed uint
	lnPort      int
	err         error
	recvBuf     []byte
)

var (
	testDialListenData = []byte("Hello tinygo :)")
)

func TestDialListen() {
	// listen thread
	listenReady := make(chan bool, 1)
	go func() {
		ln, err := net.Listen("tcp4", ":0")
		if err != nil {
			println("error listening: ", err)
			return
		}
		lnPort = ln.Addr().(*net.TCPAddr).Port

		listenReady <- true
		conn, err := ln.Accept()
		if err != nil {
			println("error accepting:", err)
			return
		}

		recvBuf = make([]byte, len(testDialListenData))
		if _, err := io.ReadFull(conn, recvBuf); err != nil {
			println("error reading: ", err)
			return
		}

		if string(recvBuf) != string(testDialListenData) {
			println("error: received data does not match sent data", string(recvBuf), " != ", string(testDialListenData))
			return
		}
		conn.Close()

		return
	}()

	<-listenReady
	conn, err := net.Dial("tcp4", "127.0.0.1:"+strconv.FormatInt(int64(lnPort), 10))
	if err != nil {
		println("error dialing: ", err)
		return
	}

	if _, err = conn.Write(testDialListenData); err != nil {
		println("error writing: ", err)
		return
	}
}

func main() {
	println("test: net start")
	TestDialListen()
	println("test: net end")
}
