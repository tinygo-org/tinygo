package main

import (
	"embed"
	"strings"
)

//go:embed a hello.txt
var files embed.FS

var (
	//go:embed "hello.*"
	helloString string

	//go:embed hello.txt
	helloBytes []byte
)

// A test to check that hidden files are not included when matching a directory.
//go:embed a/b/.hidden
var hidden string

func main() {
	println("string:", strings.TrimSpace(helloString))
	println("bytes:", strings.TrimSpace(string(helloBytes)))
	println("files:")
	readFiles(".")
}

func readFiles(dir string) {
	entries, err := files.ReadDir(dir)
	if err != nil {
		println(err.Error())
		return
	}
	for _, entry := range entries {
		entryPath := entry.Name()
		if dir != "." {
			entryPath = dir + "/" + entryPath
		}
		println("-", entryPath)
		if entry.IsDir() {
			readFiles(entryPath)
		}
	}
}
