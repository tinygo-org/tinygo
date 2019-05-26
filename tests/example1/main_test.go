package main

import (
	"testing" // This is the tinygo testing package
)

func TestFoo(t *testing.T) {
	t.Error("TestFoo failed because of stuff and things")
}
