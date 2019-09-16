package runtime

// This file implements functionality that acts as a bridge between interrupts and the scheduler.

// a double-buffered linked queue used to transfer awoken tasks from the interrupt handler to the scheduler
var wakeupQueue [2]*task
var wakeupQueueSelect bool

// previous interrupt; used to handle nesting of interrupts
var prevInt *task

// pushInt pushes a task onto the scheduler's interrupt wakeup queue.
// This is meant to be called from an interrupt to wake up a task.
// In order for this to work properly, a reference to the task must be held elsewhere.
func pushInt(t *task) {
	// interrupts before reading prevInt do not matter

	prev := prevInt // this load has to be atomic for this to work correctly

	// interrupt point A: if something interrupts here then prev.state().next will be non-nil

	prevInt = t // this store has to be atomic for this to work correctly

	if prev != nil {
		// if we were interrupted at point A, then prev.next will have a chain - so find the tail of this chain
		// nothing past prev will be modified if this is interrupted
		tail := prev
		for tail.state().next != nil {
			tail = tail.state().next
		}
		tail.state().next = t
	} else {
		// we are at the lowest level - store the base of this chain into the wakeup queue
		// we can safely be interrupted during this - any interuptors will simply place their tasks underneath t

		// select the wakeup queue not currently being accessed by the scheduler
		var wq **task
		if wakeupQueueSelect {
			wq = &wakeupQueue[1]
		} else {
			wq = &wakeupQueue[0]
		}

		// store interrupt chain into queue
		if *wq == nil {
			*wq = t
		} else {
			(*wq).state().next, *wq = *wq, t
		}
	}

	prevInt = prev
}

// popInt pops a chain of tasks off of the interrupt wakeup queue.
// This must be called twice before letting the CPU go to sleep.
func popInt() *task {
	// toggle buffers, so any interrupts that fire now write into the opposite buffer
	wakeupQueueSelect = !wakeupQueueSelect

	// get task chain
	var wq **task
	if wakeupQueueSelect {
		wq = &wakeupQueue[0]
	} else {
		wq = &wakeupQueue[1]
	}
	chain := *wq
	*wq = nil

	return chain
}
