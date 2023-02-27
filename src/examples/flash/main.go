package main

import (
	"machine"
	"time"
)

var (
	err     error
	message = "1234567887654321123456788765432112345678876543211234567887654321"
)

func main() {
	time.Sleep(3 * time.Second)

	// Print out general information
	println("Flash data start:      ", machine.FlashDataStart())
	println("Flash data end:        ", machine.FlashDataEnd())
	println("Flash data size, bytes:", machine.Flash.Size())
	println("Flash write block size:", machine.Flash.WriteBlockSize())
	println("Flash erase block size:", machine.Flash.EraseBlockSize())
	println()

	flash := machine.OpenFlashBuffer(machine.Flash, machine.FlashDataStart())
	original := make([]byte, len(message))
	saved := make([]byte, len(message))

	// Read flash contents on start (data shall survive power off)
	print("Reading data from flash: ")
	_, err = flash.Read(original)
	checkError(err)
	println(string(original))

	// Write the message to flash
	print("Writing data to flash: ")
	flash.Seek(0, 0) // rewind back to beginning
	_, err = flash.Write([]byte(message))
	checkError(err)
	println(string(message))

	// Read back flash contents after write (verify data is the same as written)
	print("Reading data back from flash: ")
	flash.Seek(0, 0) // rewind back to beginning
	_, err = flash.Read(saved)
	checkError(err)
	println(string(saved))
	println()
}

func checkError(err error) {
	if err != nil {
		for {
			println(err.Error())
			time.Sleep(time.Second)
		}
	}
}
