package main

import (
	"fmt"
	"os"
	"testing" // This is the tinygo testing package
)

func TestFoo(t *testing.T) {
	t.Error("test failed\n")
}

// TODO: change signature to accept a prepopulated test suite
func TestMain(m *testing.M) {
	fmt.Println("running tests...")
	os.Exit(m.Run())
}
