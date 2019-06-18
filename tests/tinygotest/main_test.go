package main

import (
	"testing" // This is the tinygo testing package
)

func TestFail1(t *testing.T) {
	t.Error("TestFail1 failed because of stuff and things")
}

func TestFail2(t *testing.T) {
	t.Error("TestFail2 failed for reasons")
}

func TestPass(t *testing.T) {
}
