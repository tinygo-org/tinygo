//go:build darwin || (linux && !baremetal && !tinygo.wasm && !nintendoswitch)

package runtime

// Hook dispatch for the linux/amd64 netpoller (poll_epoll_linux.go). On
// targets where that file is not compiled, netpollWaitHook is nil and the
// original futex-based behaviour is used unchanged.

// netpollWaitHook, when non-nil, is called by waitForEvents/sleepTicks with the
// maximum time the scheduler is willing to block (milliseconds, -1 = forever).
// It returns true if it blocked and handled events.
var netpollWaitHook func(timeoutMs int32) bool
