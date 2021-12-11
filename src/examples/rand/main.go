package main

import (
	"crypto/rand"
	"encoding/hex"
	"time"
)

func main() {
	var result [32]byte
	for {
		rand.Read(result[:])
		encodedString := hex.EncodeToString(result[:])
		println(encodedString)
		time.Sleep(time.Second)
	}
}
