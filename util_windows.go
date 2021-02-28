package main

// This file contains utility functions for Windows.

import (
	"os/exec"
	"syscall"

	"golang.org/x/sys/windows"
)

// setCommandAsDaemon makes sure this command does not receive signals sent to
// the parent.
func setCommandAsDaemon(daemon *exec.Cmd) {
	daemon.SysProcAttr = &syscall.SysProcAttr{
		CreationFlags: windows.DETACHED_PROCESS,
	}
}
