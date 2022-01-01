package main

import (
	"testing" // This is the tinygo testing package
)

func TestFail1(t *testing.T) {
	t.Error("TestFail1 failed because of stuff and things")
}

func TestFail2(t *testing.T) {
	t.Fatalf("TestFail2 failed for %v ", "reasons")
}

func TestFail3(t *testing.T) {
	t.Fail()
	t.Logf("TestFail3 failed for %v ", "reasons")
}

func TestPass(t *testing.T) {
	t.Log("TestPass passed")
}

func TestSkip(t *testing.T) {
	t.Skip("skip this")

	panic("test was not skipped")
}

func BenchmarkNotImplemented(b *testing.B) {
}
