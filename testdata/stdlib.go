package main

import (
	"fmt"
	"math/rand"
	"os"
)

func main() {
	fmt.Println("stdin: ", os.Stdin.Fd())
	fmt.Println("stdout:", os.Stdout.Fd())
	fmt.Println("stderr:", os.Stderr.Fd())

	fmt.Println("pseudorandom number:", rand.Int31())
}
