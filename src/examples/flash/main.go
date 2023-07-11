package main

import (
	"machine"
	"time"
)

var (
	err     error
	message = "1234567887654321123456788765432112345678876543211234567887654321" +
		"1234567887654321123456788765432112345678876543211234567887654321" +
		"1234567887654321123456788765432112345678876543211234567887654321" +
		"1234567887654321123456788765432112345678876543211234567887654321"
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

	original := make([]byte, len(message))
	saved := make([]byte, len(message))

	// Read flash contents on start (data shall survive power off)
	println("Reading original data from flash:")
	_, err = machine.Flash.ReadAt(original, 0)
	checkError(err)
	println(string(original))

	// erase flash
	println("Erasing flash...")
	needed := int64(len(message)) / machine.Flash.EraseBlockSize()
	if needed == 0 {
		// have to erase at least 1 block
		needed = 1
	}

	err := machine.Flash.EraseBlocks(0, needed)
	checkError(err)

	// Write the message to flash
	println("Writing new data to flash:")
	_, err = machine.Flash.WriteAt([]byte(message), 0)
	checkError(err)
	println(string(message))

	// Read back flash contents after write (verify data is the same as written)
	println("Reading data back from flash: ")
	_, err = machine.Flash.ReadAt(saved, 0)
	checkError(err)
	if !equal(saved, []byte(message)) {
		println("data verify error")
	}

	println(string(saved))
	println("Done.")
}

func equal(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	for i, v := range a {
		if v != b[i] {
			return false
		}
	}
	return true
}

func checkError(err error) {
	if err != nil {
		for {
			println(err.Error())
			time.Sleep(time.Second)
		}
	}
}
