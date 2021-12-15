package main

import (
	"fmt"
	"math/rand"
	"os"
	"strings"
	"syscall"
	"time"
)

func main() {
	// package os, fmt
	fmt.Println("stdin: ", os.Stdin.Name())
	fmt.Println("stdout:", os.Stdout.Name())
	fmt.Println("stderr:", os.Stderr.Name())

	// Package syscall, this mostly checks whether the calls don't trigger an error.
	syscall.Getuid()
	syscall.Geteuid()
	syscall.Getgid()
	syscall.Getegid()
	syscall.Getpid()
	syscall.Getppid()

	// package math/rand
	fmt.Println("pseudorandom number:", rand.Int31())

	// package strings
	fmt.Println("strings.IndexByte:", strings.IndexByte("asdf", 'd'))
	fmt.Println("strings.Replace:", strings.Replace("An example string", " ", "-", -1))

	// package time
	time.Sleep(time.Millisecond)
	time.Sleep(-1) // negative sleep should return immediately

	// Exit the program normally.
	os.Exit(0)
}
