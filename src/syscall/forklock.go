//go:build tinygo && linux && !wasip1 && !wasip2 && tinygo.wasm && !wasm_unknown && !darwin && !baremetal && !nintendoswitch

package syscall

import (
	"sync"
)

var ForkLock sync.RWMutex

func CloseOnExec(fd int) {
	system.CloseOnExec(fd)
}

func SetNonblock(fd int, nonblocking bool) (err error) {
	return system.SetNonblock(fd, nonblocking)
}
