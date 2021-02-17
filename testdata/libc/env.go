package main

import (
	"os"
)

func main() {
	println(os.Getenv("ENV1"))
	v, ok := os.LookupEnv("ENV2")
	if !ok {
		panic("ENV2 not found")
	}
	println(v)
}
