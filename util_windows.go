package main

// This file contains utility functions for Windows.

import (
	"os/exec"
)

// setCommandAsDaemon makes sure this command does not receive signals sent to
// the parent. It is unimplemented on Windows.
func setCommandAsDaemon(daemon *exec.Cmd) {
	// Not implemented.
}
