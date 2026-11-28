package main

import (
	"bufio"
	"os"
)

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	// A UART has no end of file, so an empty line also ends the input.
	for scanner.Scan() && scanner.Text() != "" {
		println("line:", scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		println("error:", err.Error())
	}
	println("done")
}
