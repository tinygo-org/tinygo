package main

import (
	"mime/multipart"
	"strings"
)

func main() {
	r := multipart.NewReader(strings.NewReader("--x\r\n\r\n"), "x")
	_, err := r.NextPart()
	println(err != nil)
}
