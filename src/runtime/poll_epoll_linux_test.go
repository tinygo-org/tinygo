//go:build linux && amd64 && !baremetal && !tinygo.wasm && !nintendoswitch

package runtime

import "testing"

// TestNetpollWaitHookInstalled checks that the linux/amd64 epoll netpoller
// installs its scheduler hook (netpollWaitHook), so waitForEvents/sleepTicks
// route through epoll_wait when fds are registered.
func TestNetpollWaitHookInstalled(t *testing.T) {
	if netpollWaitHook == nil {
		t.Fatal("netpollWaitHook is nil: the epoll netpoller is not installed")
	}
}

// TestNetpollWaitHookNoFds checks the hook declines to block when no fd is
// registered with the poller, leaving the stock futex wait in place.
func TestNetpollWaitHookNoFds(t *testing.T) {
	if netpollWaitHook == nil {
		t.Fatal("netpollWaitHook is nil")
	}
	if netpollWaitHook(-1) {
		t.Error("netpollWaitHook(-1) = true with no registered fds; want false (stock behaviour)")
	}
}
