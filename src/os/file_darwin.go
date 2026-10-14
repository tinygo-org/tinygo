package os

import "syscall"

func pipe(p []int) error {
	// Darwin has no pipe2, so mark the descriptors close-on-exec afterwards.
	// ForkLock keeps a spawn out of the window between the two steps.
	syscall.ForkLock.RLock()
	defer syscall.ForkLock.RUnlock()
	if err := syscall.Pipe(p); err != nil {
		return err
	}
	syscall.CloseOnExec(p[0])
	syscall.CloseOnExec(p[1])
	return nil
}
