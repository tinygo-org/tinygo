//go:build !windows
// +build !windows

package main

// This file contains utility functions for Unix-like systems (e.g. Linux).

import (
	"os/exec"
	"syscall"
)

// setCommandAsDaemon makes sure this command does not receive signals sent to
// the parent.
func setCommandAsDaemon(daemon *exec.Cmd) {
	// https://stackoverflow.com/a/35435038/559350
	daemon.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
		Pgid:    0,
	}
}
