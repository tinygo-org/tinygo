package sync

import "internal/task"

type WaitGroup struct {
	counter uint
	waiters task.Stack
}

func (wg *WaitGroup) Add(delta int) {
	if delta > 0 {
		// Check for overflow.
		if uint(delta) > (^uint(0))-wg.counter {
			panic("sync: WaitGroup counter overflowed")
		}

		// Add to the counter.
		wg.counter += uint(delta)
	} else {
		// Check for underflow.
		if uint(-delta) > wg.counter {
			panic("sync: negative WaitGroup counter")
		}

		// Subtract from the counter.
		wg.counter -= uint(-delta)

		// If the counter is zero, everything is done and the waiters should be resumed.
		// This code assumes that the waiters cannot wake up until after this function returns.
		// In the current implementation, this is always correct.
		if wg.counter == 0 {
			for t := wg.waiters.Pop(); t != nil; t = wg.waiters.Pop() {
				scheduleTask(t)
			}
		}
	}
}

func (wg *WaitGroup) Done() {
	wg.Add(-1)
}

func (wg *WaitGroup) Wait() {
	if wg.counter == 0 {
		// Everything already finished.
		return
	}

	// Push the current goroutine onto the waiter stack.
	wg.waiters.Push(task.Current())

	// Pause until the waiters are awoken by Add/Done.
	task.Pause()
}
