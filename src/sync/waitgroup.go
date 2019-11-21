package sync

type WaitGroup struct {
	count int
	done  chan struct{}
}

func (wg *WaitGroup) Add(delta int) {
	if delta < 0 && wg.count < -delta {
		panic("sync: negative WaitGroup counter")
	}
	wg.count += delta
	if wg.count == 0 && wg.done != nil {
		close(wg.done)
		wg.done = nil
	}
}

func (wg *WaitGroup) Done() {
	wg.Add(-1)
}

func (wg *WaitGroup) Wait() {
	if wg.count == 0 {
		// already done
		return
	}
	if wg.done == nil {
		// first waiter, initialize channel
		wg.done = make(chan struct{})
	}

	// wait for completion notification, signalled by channel closure
	<-wg.done
}
