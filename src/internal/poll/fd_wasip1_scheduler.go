//go:build wasip1 && (scheduler.tasks || scheduler.asyncify)

package poll

import (
	"internal/task"
	"time"
)

//go:linkname runtime_netpoll_addwait runtime.runtime_netpoll_addwait
func runtime_netpoll_addwait(fd uint32, mode uint8) uintptr

//go:linkname runtime_netpoll_done runtime.runtime_netpoll_done
func runtime_netpoll_done(pd uintptr)

//go:linkname runtime_netpoll_wake runtime.runtime_netpoll_wake
func runtime_netpoll_wake(pd uintptr)

func wait(fd int, mode uint8) error {
	pd := runtime_netpoll_addwait(uint32(fd), mode)
	task.Pause()
	runtime_netpoll_done(pd)
	return nil
}

// The timer and pollIO can both wake the same task. The pollDesc fired flag
// prevents the second wake from adding the task to the run queue again.
func (fd *FD) parkUntil(mode uint8, deadline time.Time) error {
	d := time.Until(deadline)
	if d <= 0 {
		return ErrDeadlineExceeded
	}
	pd := runtime_netpoll_addwait(uint32(fd.Sysfd), mode)
	timer := time.AfterFunc(d, func() {
		runtime_netpoll_wake(pd)
	})
	task.Pause()
	timer.Stop()
	runtime_netpoll_done(pd)
	return nil
}
