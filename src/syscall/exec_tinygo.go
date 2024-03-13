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
// So for now, make our own ForkLock that doesn't use sync..

type forklock struct {
}

func (f forklock) RLock() {
	println("ForkLock.RLock not implemented")
}

func (f forklock) RUnlock() {
	println("ForkLock.RUnlock not implemented")
}

var ForkLock forklock

func CloseOnExec(fd int) {
	println("CloseOnExec not implemented")
}

func SetNonblock(fd int, nonblocking bool) (err error) {
	println("SetNonblock not implemented", fd, nonblocking)
	return EOPNOTSUPP
}

type SysProcAttr struct {
}
