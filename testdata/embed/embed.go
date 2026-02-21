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

// A test to check that hidden files ARE included when using "all:" prefix.
//go:embed all:a
var allFiles embed.FS

var helloStringBytes = []byte(helloString)

func main() {
	println("string:", strings.TrimSpace(helloString))
	println("bytes:", strings.TrimSpace(string(helloBytes)))
	println("[]byte(string):", strings.TrimSpace(string(helloStringBytes)))
	println("files:")
	readFiles(".", files)
	println("all:a files (should include .hidden):")
	readFiles(".", allFiles)
}

func readFiles(dir string, fs embed.FS) {
	entries, err := fs.ReadDir(dir)
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
			readFiles(entryPath, fs)
		}
	}
}
