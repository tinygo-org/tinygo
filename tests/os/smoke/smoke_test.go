package os_smoke_test

// Simple smoke tests for the os package or things that depend on it.
// Intended to catch build tag mistakes affecting bare metal targets.

import (
	"crypto/rand"
	"crypto/rsa"
	"fmt"
	"path/filepath"
	"testing"
)

// Regression test for https://github.com/tinygo-org/tinygo/issues/2563
func TestFilepath(t *testing.T) {
	if filepath.Base("foo/bar") != "bar" {
		t.Errorf("filepath.Base is very confused")
	}
}

// Regression test for https://github.com/tinygo-org/tinygo/issues/2530
func TestFmt(t *testing.T) {
	n, err := fmt.Printf("Hello, world!\n")
	if err != nil {
		t.Errorf("printf returned error %s", err)
	} else if n != 14 {
		t.Errorf("printf returned %d, expected 14", n)
	}
}

// Regression test for https://github.com/tinygo-org/tinygo/issues/4921
func TestRand(t *testing.T) {
	if _, err := rsa.GenerateKey(rand.Reader, 2048); err != nil {
		t.Error(err)
	}
}
