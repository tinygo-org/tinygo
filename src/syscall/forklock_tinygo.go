//go:build tinygo

package syscall

// This is the original ForkLock:
//
// var ForkLock sync.RWMutex
//
// This requires importing sync, but importing sync causes an import loop:
//
// package tinygo.org/x/drivers/examples/net/tcpclient
//        imports bytes
//        imports io
//        imports sync
//        imports internal/task
//        imports runtime/interrupt
//        imports device/arm
//        imports syscall
//        imports sync: import cycle not allowed
//
// So for now, make our own stubbed-out ForkLock that doesn't use sync..

type forklock struct{}

func (f forklock) RLock()   {}
func (f forklock) RUnlock() {}

var ForkLock forklock

func CloseOnExec(fd int) {
	system.CloseOnExec(fd)
}

func SetNonblock(fd int, nonblocking bool) (err error) {
	return system.SetNonblock(fd, nonblocking)
}

type SysProcAttr struct{}
