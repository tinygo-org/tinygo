package main

import "crypto/rand"

// TODO: make this a test in the crypto/rand package.

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
}
