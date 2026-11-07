package main

import (
	"errors"
	"io"
	"io/fs"
	"os"
	"path"
	"runtime"
)

func init() {
	if runtime.GOOS != "wasip1" {
		return
	}
	cwd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	if pwd := os.Getenv("PWD"); pwd != "" && cwd != path.Join("/", pwd) {
		panic("initial working directory does not match PWD")
	}
	runtime.GC()
	if _, err := os.ReadFile("testdata/filesystem.txt"); err != nil {
		panic(err)
	}
}

func main() {
	_, err := os.Open("non-exist")
	if !errors.Is(err, fs.ErrNotExist) {
		panic("should be non exist error")
	}

	f, err := os.Open("testdata/filesystem.txt")
	if err != nil {
		panic(err)
	}

	defer func() {
		if err := f.Close(); err != nil {
			panic(err)
		}

		// read after close: error should be returned
		_, err := f.Read(make([]byte, 10))
		if err == nil {
			panic("error expected for reading after closing files")
		}
	}()

	data, err := io.ReadAll(f)
	if err != nil {
		panic(err)
	}

	os.Stdout.Write(data)

	cwd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	if cwd == "" {
		panic("path is empty")
	}
	if runtime.GOOS == "wasip1" {
		if err := os.Chdir("testdata"); err != nil {
			panic(err)
		}
		changed, err := os.Getwd()
		if err != nil {
			panic(err)
		}
		runtime.GC()
		data, err := os.ReadFile("filesystem.txt")
		if err != nil || len(data) == 0 {
			panic("relative read after Chdir failed")
		}
		if err := os.Chdir(cwd); err != nil {
			panic(err)
		}
		if changed != path.Join(cwd, "testdata") {
			panic("Getwd result changed after another Chdir")
		}
	}
}
