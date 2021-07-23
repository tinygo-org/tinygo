package main

import (
	"embed"
	"log"
)

//go:embed file*.txt
//go:embed index.html styles.css
var files embed.FS

func main() {
	println(".....")
	println(msg)
	contents, err := files.ReadDir(".")
	if err != nil {
		log.Fatal(err)
	}
	for _, c := range contents {
		println("file:", c.Name())
	}
}
