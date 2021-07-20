package main

import (
	"embed"
	"log"
)

//go:embed file1.txt file2.txt
//go:embed file3.txt
var files embed.FS

func main() {
	println(msg)
	contents, err := files.ReadDir(".")
	if err != nil {
		log.Fatal(err)
	}
	for _, c := range contents {
		println("file:", c.Name())
	}
}
