package runtime

// A naive implementation of coroutines that supports
// package iter.

type coro struct {
	f  func(*coro)
	ch chan struct{}
}

//go:linkname newcoro

func newcoro(f func(*coro)) *coro {
	c := &coro{
		ch: make(chan struct{}),
		f:  f,
	}
	go func() {
		defer close(c.ch)
		<-c.ch
		f(c)
	}()
	return c
}

//go:linkname coroswitch

func coroswitch(c *coro) {
	c.ch <- struct{}{}
	<-c.ch
}
