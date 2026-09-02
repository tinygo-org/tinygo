//go:build scheduler.cores

package runtime

import "internal/task"

type runtimeSpinLock struct {
	state task.Uint32
}

func (lock *runtimeSpinLock) Lock() {
	for !lock.state.CompareAndSwap(0, 1) {
	}
}

func (lock *runtimeSpinLock) Unlock() {
	lock.state.Store(0)
}

type timerLock = runtimeSpinLock
type synctestLock = runtimeSpinLock
