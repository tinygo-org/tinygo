package main

import (
	"machine"
	"strconv"
	"time"
)

func main() {
	time.Sleep(2 * time.Second) // connect via serial
	buf1 := makeBuffer('|', 600)
	buf2 := makeBuffer('/', 600)
	println("start")
	serialWrite(buf1)
	serialWrite(buf2)
}

func makeBuffer(sep byte, size int) []byte {
	buf := make([]byte, size)
	for i := 0; i < size-5; i += 5 {
		buf[i] = sep
		strconv.AppendInt(buf[i+1:i+1:i+5], int64(i), 10)
	}
	return buf
}

func serialWrite(b []byte) {
	machine.Serial.Write(b)
}
