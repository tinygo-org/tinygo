package main

import (
	"bufio"
	"os"
)

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		println("line:", scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		println("error:", err.Error())
	}
	println("done")
}
