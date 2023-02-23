package main

import (
	"machine"
	"time"
)

var (
	err  error
	data = "1234567887654321123456788765432112345678876543211234567887654321"
)

func main() {
	time.Sleep(time.Second)

	println("flash data start:", machine.FlashDataStart())
	println("flash data end:  ", machine.FlashDataEnd())

	println("writing flash data...")
	buf := machine.OpenFlashBuffer(machine.Flash, machine.FlashDataStart())
	_, err = buf.Write([]byte(data))
	if err != nil {
		for {
			println(err.Error())
			time.Sleep(time.Second)
		}
	}

	// rewind back to beginning
	buf.Seek(0, 0)

	result := make([]byte, len(data))
	_, err = buf.Read(result)
	if err != nil {
		for {
			println(err.Error())
			time.Sleep(time.Second)
		}
	}

	for {
		println("result:", string(result))
		time.Sleep(time.Second)
	}
}
