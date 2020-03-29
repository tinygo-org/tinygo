package runtime

import (
	"internal/task"
	"runtime/volatile"
)

// timer is an interface which can be implemented as a wrapper around a timer interrupt.
type timer interface {
	// setTimer configures the timer interrupt to fire when the ticks() value reaches the specified wakeup time.
	// The timer interrupt must be disabled before calling this function.
	setTimer(wakeup timeUnit)

	// disableTimer disables the timer interrupt.
	disableTimer()
}

// timerController is a type which manages a hardware timer, with builtin queueing.
type timerController struct {
	// fired is used to indicate whether the timer interrupt handler fired:
	//   0: the timer interrupt handler has not fired
	//   1: the timer interrupt handler fired
	fired volatile.Register8

	t timer

	baseTime timeUnit
	queue    *task.Task
}

// set configures the timer interrupt to fire when the next timer expires.
// This function assumes that there is a timer on the queue.
func (c *timerController) set() {
	// Disable the timer to avoid a race condition.
	c.t.disableTimer()

	// Clear the fired flag.
	c.fired.Set(0)

	// Set the timer with the appropriate time.
	c.t.setTimer(c.baseTime)
}

// enqueue queues a task to be awoken at the specified wakeup time.
func (c *timerController) enqueue(t *task.Task, wakeup timeUnit) {
	t.Data = 0

	// Insert task into timer queue.
	var baseUpdated bool
	switch {
	case c.queue == nil:
		// There were no timers previously in the queue.
		scheduleLog("  new timer queue")
		c.queue = t
		c.baseTime = wakeup
		baseUpdated = true
	case wakeup < c.baseTime:
		// This task comes before everything else on the queue.
		c.queue.Data += uint(c.baseTime - wakeup)
		t.Next = c.queue
		c.queue = t
		c.baseTime = wakeup
		baseUpdated = true
	default:
		// Insert this task later in the queue.
		offset := wakeup - c.baseTime
		prev := c.queue
		for ; prev.Next != nil && offset < timeUnit(prev.Next.Data); prev = prev.Next {
			offset -= timeUnit(prev.Data)
		}
		prev.Next, t.Next = t, prev.Next
		if t.Next != nil {
			t.Next.Data -= uint(offset)
		}
		t.Data = uint(offset)
	}

	if c.fired.Get() != 0 {
		// The timer already fired.
		// Poll for completed timer events.
		c.poll()
		return
	}

	if baseUpdated {
		// The base time has changed, so the timer interrupt needs to be reconfigured.
		c.set()
	}
}

// poll checks for any expired timers and schedules them.
// This must be called periodically by the scheduler.
func (c *timerController) poll() bool {
	if c.fired.Get() == 0 {
		// The timer has not fired.
		return false
	}

	// Clear fired flag.
	c.fired.Set(0)

	if c.queue == nil {
		// The timer fired even though we did not ask it to.
		scheduleLog("  unrequested timer fired")
		return false
	}

	// The first timer is known to be complete.
	{
		t := c.queue
		c.queue = t.Next
		t.Next = nil
		runqueue.Push(t)
	}

	if c.queue == nil {
		// Bail out early to avoid the call to ticks.
		scheduleLog("  all timers have expired")
		return true
	}

	// Check the time.
	now := ticks()

	// Pop and schedule all tasks with expired timers.
	for c.queue != nil && c.baseTime+timeUnit(c.queue.Data) < now {
		// Pop a task from the queue.
		t := c.queue.Next
		c.queue = t.Next
		t.Next = nil

		// Adjust base time.
		c.baseTime += timeUnit(t.Data)

		// Schedule task.
		runqueue.Push(t)
	}

	if c.queue == nil {
		scheduleLog("  all timers have expired")
		return true
	}

	// Normalize representation of queue such that the first task has a zero offset.
	c.baseTime += timeUnit(c.queue.Data)
	c.queue.Data = 0

	// Configure the interrupt.
	c.set()

	return true
}

func (c *timerController) handleInterrupt() {
	c.t.disableTimer()
	c.fired.Set(1)
}
