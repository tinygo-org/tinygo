package main

import (
	"fmt"
	"math/rand"
	"os"
	"strings"
)

func main() {
	// package os, fmt
	fmt.Println("stdin: ", os.Stdin.Fd())
	fmt.Println("stdout:", os.Stdout.Fd())
	fmt.Println("stderr:", os.Stderr.Fd())

	// package math/rand
	fmt.Println("pseudorandom number:", rand.Int31())

	// package strings
	fmt.Println("strings.IndexByte:", strings.IndexByte("asdf", 'd'))
}
