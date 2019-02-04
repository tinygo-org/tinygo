package main

import (
	"fmt"
	"os"
)

func main() {
	fmt.Println("stdin: ", os.Stdin.Fd())
	fmt.Println("stdout:", os.Stdout.Fd())
	fmt.Println("stderr:", os.Stderr.Fd())
}
