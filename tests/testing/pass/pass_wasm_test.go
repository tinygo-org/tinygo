//go:build wasm

package pass_test

import "testing"

func TestDeepSuspendWithDefer(t *testing.T) {
	ready := make(chan struct{})
	release := make(chan struct{})
	done := make(chan struct{})
	go func() {
		deepSuspendWithDefer(512, ready, release)
		close(done)
	}()
	<-ready
	close(release)
	<-done
}

func deepSuspendWithDefer(depth int, ready chan<- struct{}, release <-chan struct{}) {
	defer func() {}()
	if depth == 0 {
		ready <- struct{}{}
		<-release
		return
	}
	deepSuspendWithDefer(depth-1, ready, release)
}
