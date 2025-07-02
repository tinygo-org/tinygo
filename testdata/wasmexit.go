package main

import (
	"os"
	"time"
)

func main() {
	println("wasmexit test:", os.Args[1])
	switch os.Args[1] {
	case "normal":
		return
	case "exit-0":
		os.Exit(0)
	case "exit-0-sleep":
		time.Sleep(time.Millisecond)
		println("slept")
		os.Exit(0)
	case "exit-1":
		os.Exit(1)
	case "exit-1-sleep":
		time.Sleep(time.Millisecond)
		println("slept")
		os.Exit(1)
	}
	println("unknown wasmexit test")
}
