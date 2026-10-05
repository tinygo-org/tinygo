package main

// This file contains utility functions for Windows.

import (
	"os"
	"os/exec"
	"syscall"

	"golang.org/x/sys/windows"
)

func resolveDirectory(name string) (string, error) {
	dir, err := os.Open(name)
	if err != nil {
		return "", err
	}
	defer dir.Close()

	// Resolve junctions, which filepath.EvalSymlinks does not follow.
	// https://learn.microsoft.com/windows/win32/api/fileapi/nf-fileapi-getfinalpathnamebyhandlew
	buffer := make([]uint16, windows.MAX_PATH+1)
	for {
		n, err := windows.GetFinalPathNameByHandle(windows.Handle(dir.Fd()), &buffer[0], uint32(len(buffer)), 0)
		if err != nil {
			return "", err
		}
		if n < uint32(len(buffer)) {
			return windows.UTF16ToString(buffer[:n]), nil
		}
		buffer = make([]uint16, n+1)
	}
}

// setCommandAsDaemon makes sure this command does not receive signals sent to
// the parent.
func setCommandAsDaemon(daemon *exec.Cmd) {
	daemon.SysProcAttr = &syscall.SysProcAttr{
		CreationFlags: windows.DETACHED_PROCESS,
	}
}
