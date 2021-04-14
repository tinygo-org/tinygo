package main

import (
	"fmt"
	"math/rand"
	"os"
	"strings"
)

func main() {
	// package os, fmt
	fmt.Println("stdin: ", os.Stdin.Name())
	fmt.Println("stdout:", os.Stdout.Name())
	fmt.Println("stderr:", os.Stderr.Name())

	// package math/rand
	fmt.Println("pseudorandom number:", rand.Int31())

	// package strings
	fmt.Println("strings.IndexByte:", strings.IndexByte("asdf", 'd'))
	fmt.Println("strings.Replace:", strings.Replace("An example string", " ", "-", -1))

	// Exit the program normally.
	os.Exit(0)
}
