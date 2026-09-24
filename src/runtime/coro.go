package runtime

import "internal/task"

// A naive implementation of coroutines that supports
// package iter.

type coro struct {
	f    func(*coro)
	task *task.Task
}

//go:linkname newcoro

func newcoro(f func(*coro)) *coro {
	c := &coro{f: f}
	ready := make(chan struct{})
	go func() {
		current := task.Current()
		c.task = current
		inBubble := synctestTaskBlockBegin(current)
		synctestTaskBlock(current)
		close(ready)
		if inBubble {
			synctestTaskBlockEnd(current, false)
		}
		task.Pause()
		defer coroexit(c)
		f(c)
	}()
	<-ready
	return c
}

func coroexit(c *coro) {
	task.CoroExit(c.task)
}

//go:linkname coroswitch

func coroswitch(c *coro) {
	current := task.Current()
	next := c.task
	c.task = current
	inBubble := synctestTaskBlockBegin(current)
	synctestTaskBlock(current)
	scheduleTask(next)
	if inBubble {
		synctestTaskBlockEnd(current, false)
	}
	task.Pause()
}
