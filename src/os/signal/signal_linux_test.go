//go:build !plan9 && !windows && !js && !wasm && !wasip1 && !wasip2
// +build !plan9,!windows,!js,!wasm,!wasip1,!wasip2

package signal

import (
	"runtime"
	"testing"
)

// The file sigqueue.go contains preliminary stubs for signal handling.
// Since tinygo ultimately lacks support for signals, these stubs are
// placeholders for future work.
// There might be userland applications that rely on these functions
// from the upstream go package os/signal that we want to enable
// building and linking.

func TestSignalHandling(t *testing.T) {
	if runtime.GOOS == "wasip1" || runtime.GOOS == "wasip2" {
		t.Skip()
	}

	// This test is a placeholder for future work.
	// It is here to ensure that the stubs in sigqueue.go
	// are correctly linked and can be called from userland
	// applications.
	enableSignal(0)
	disableSignal(0)
	ignoreSignal(0)
}
