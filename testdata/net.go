package main

import (
	"bytes"
	"fmt"
	"net"
	"strconv"
	"time"
)

// Test golang network package integration for tinygo.
// This test is not exhaustive and only tests the basic functionality of the package.

const (
	TEST_PORT = 9000
)

var (
	testsPassed uint
	err         error
	sendBuf     = &bytes.Buffer{}
	recvBuf     = &bytes.Buffer{}
)

var (
	testDialListenData = []byte("Hello tinygo :)")
)

func TestDialListen() {
	// listen thread
	go func() {
		ln, err := net.Listen("tcp", ":"+strconv.FormatInt(TEST_PORT, 10))
		if err != nil {
			fmt.Printf("error listening: %v\n", err)
			return
		}

		conn, err := ln.Accept()
		if err != nil {
			fmt.Printf("error accepting: %v\n", err)
			return
		}

		recvBuf.Reset()
		_, err = conn.Read(recvBuf.Bytes())
		if err != nil {
			fmt.Printf("error reading: %v\n", err)
			return
		}

		// TODO: this is racy
		if recvBuf.String() != string(testDialListenData) {
			fmt.Printf("error: received data does not match sent data: '%s' != '%s'\n", recvBuf.String(), string(testDialListenData))
			return
		}
		conn.Close()

		return
	}()

	// hacky way to wait for the listener to start
	time.Sleep(1 * time.Second)

	sendBuf.Reset()
	fmt.Fprint(sendBuf, testDialListenData)
	conn, err := net.Dial("tcp4", "127.0.0.1:"+strconv.FormatInt(TEST_PORT, 10))
	if err != nil {
		fmt.Printf("error dialing: %v\n", err)
		return
	}

	if _, err = conn.Write(sendBuf.Bytes()); err != nil {
		fmt.Printf("error writing: %v\n", err)
		return
	}
}

func main() {
	fmt.Printf("test: net start\n")
	TestDialListen()
	fmt.Printf("test: net end\n")
}
