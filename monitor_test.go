package main

import (
	"bytes"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/tinygo-org/tinygo/builder"
	"github.com/tinygo-org/tinygo/compileopts"
)

func TestTraceback(t *testing.T) {
	if runtime.GOOS != "linux" {
		// We care about testing the ELF format, which is only used on Linux
		// (not on MacOS or Windows).
		t.Skip("Test only works on Linux")
	}

	// Build a small binary that only panics.
	tmpdir := t.TempDir()
	config, err := builder.NewConfig(&compileopts.Options{
		GOOS:          runtime.GOOS,
		GOARCH:        runtime.GOARCH,
		Opt:           "z",
		InterpTimeout: time.Minute,
		Debug:         true,
	})
	if err != nil {
		t.Fatal(err)
	}
	result, err := builder.Build("testdata/trivialpanic.go", ".elf", tmpdir, config)
	if err != nil {
		t.Fatal(err)
	}

	// Run this binary, and capture the output.
	buf := &bytes.Buffer{}
	cmd := exec.Command(result.Binary)
	cmd.Stdout = buf
	cmd.Stderr = buf
	cmd.Run() // this will return an error because of the panic, ignore it

	// Extract the "runtime error at" address.
	line := bytes.TrimSpace(buf.Bytes())
	address := extractPanicAddress(line)
	if address == 0 {
		t.Fatalf("could not extract panic address from %#v", string(line))
	}

	// Look up the source location for this address.
	location, err := addressToLine(result.Executable, address)
	if err != nil {
		t.Fatal("could not read source location:", err)
	}

	// Verify that the source location is as expected.
	if filepath.Base(location.Filename) != "trivialpanic.go" {
		t.Errorf("expected path to end with trivialpanic.go, got %#v", location.Filename)
	}
	if location.Line != 6 {
		t.Errorf("expected panic location to be line 6, got line %d", location.Line)
	}
}
