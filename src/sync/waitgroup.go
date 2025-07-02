package sync

import "internal/task"

type WaitGroup struct {
	futex task.Futex
}

func (wg *WaitGroup) Add(delta int) {
	switch {
	case delta > 0:
		// Delta is positive.
		for {
			// Check for overflow.
			counter := wg.futex.Load()
			if uint32(delta) > (^uint32(0))-counter {
				panic("sync: WaitGroup counter overflowed")
			}

			// Add to the counter.
			if wg.futex.CompareAndSwap(counter, counter+uint32(delta)) {
				// Successfully added.
				return
			}
		}
	default:
		// Delta is negative (or zero).
		for {
			counter := wg.futex.Load()

			// Check for underflow.
			if uint32(-delta) > counter {
				panic("sync: negative WaitGroup counter")
			}

			// Subtract from the counter.
			if !wg.futex.CompareAndSwap(counter, counter-uint32(-delta)) {
				// Could not swap, trying again.
				continue
			}

			// If the counter is zero, everything is done and the waiters should
			// be resumed.
			// When there are multiple thread, there is a chance for the counter
			// to go to zero, WakeAll to be called, and then the counter to be
			// incremented again before a waiting goroutine has a chance to
			// check the new (zero) value. However the last increment is
			// explicitly given in the docs as something that should not be
			// done:
			//
			//   > Note that calls with a positive delta that occur when the
			//   > counter is zero must happen before a Wait.
			//
			// So we're fine here.
			if counter-uint32(-delta) == 0 {
				// TODO: this is not the most efficient implementation possible
				// because we wake up all waiters unconditionally, even if there
				// might be none. Though since the common usage is for this to
				// be called with at least one waiter, it's probably fine.
				wg.futex.WakeAll()
			}

			// Successfully swapped (and woken all waiting tasks if needed).
			return
		}
	}
}

func (wg *WaitGroup) Done() {
	wg.Add(-1)
}

func (wg *WaitGroup) Wait() {
	for {
		counter := wg.futex.Load()
		if counter == 0 {
			return // everything already finished
		}

		if wg.futex.Wait(counter) {
			// Successfully woken by WakeAll (in wg.Add).
			break
		}
	}
}
