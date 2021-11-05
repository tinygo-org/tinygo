package main

import (
	"io/ioutil"
	"os"
)

func main() {
	_, err := os.Open("non-exist")
	if !os.IsNotExist(err) {
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

	data, err := ioutil.ReadAll(f)
	if err != nil {
		panic(err)
	}

	os.Stdout.Write(data)
}
