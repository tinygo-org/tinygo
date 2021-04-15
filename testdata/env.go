package main

import (
	"os"
)

func main() {
	println("ENV1:", os.Getenv("ENV1"))
	v, ok := os.LookupEnv("ENV2")
	if !ok {
		println("ENV2 not found")
	}
	println("ENV2:", v)
}
