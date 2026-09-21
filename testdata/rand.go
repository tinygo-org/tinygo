package main

import (
	"crypto/rand"
	"errors"
)

// TODO: make this a test in the crypto/rand package.

type failingReader struct{}

func (failingReader) Read([]byte) (int, error) {
	return 0, errors.New("test failure")
}

func main() {
	buf := make([]byte, 500)
	n, err := rand.Read(buf)
	if n != len(buf) || err != nil {
		println("could not read random numbers:", err)
	}

	// Very simple test that random numbers are at least somewhat random.
	sum := 0
	for _, b := range buf {
		sum += int(b)
	}
	if sum < 95*len(buf) || sum > 159*len(buf) {
		println("random numbers don't seem that random, the average byte is", sum/len(buf))
	} else {
		println("random number check was successful")
	}

	text := rand.Text()
	valid := len(text) == 26
	for _, c := range text {
		if (c < 'A' || c > 'Z') && (c < '2' || c > '7') {
			valid = false
		}
	}
	if !valid {
		println("random text is invalid:", text)
	} else {
		println("random text check was successful")
	}

	reader := rand.Reader
	rand.Reader = failingReader{}
	func() {
		defer func() {
			if value := recover(); value != "crypto/rand: failed to read random data: test failure" {
				println("unexpected panic from random text")
			} else {
				println("random text read failure was caught")
			}
		}()
		rand.Text()
	}()
	rand.Reader = reader
}
