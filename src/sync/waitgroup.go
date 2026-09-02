package sync

import (
	"internal/task"
	"unsafe"
)

type WaitGroup struct {
	futex    task.Futex
	lock     task.PMutex
	waiters  task.Stack
	counter  int
	waiting  int
	synctest unsafe.Pointer
}

func (wg *WaitGroup) Add(delta int) {
	if !synctestIsEnabled() {
		wg.addPlain(delta)
		return
	}

	currentBubble := task.Current().SynctestBubble
	wg.lock.Lock()
	if currentBubble == nil && wg.synctest == nil {
		wg.lock.Unlock()
		wg.addPlain(delta)
		return
	}
	if wg.synctest == nil && wg.futex.Load() != 0 {
		wg.lock.Unlock()
		runtimeFatal("sync: WaitGroup.Add called from inside and outside synctest bubble")
	}
	if currentBubble != nil {
		if wg.synctest == nil {
			wg.synctest = currentBubble
		} else if wg.synctest != currentBubble {
			wg.lock.Unlock()
			runtimeFatal("sync: WaitGroup.Add called from multiple synctest bubbles")
		}
	} else if wg.synctest != nil {
		wg.lock.Unlock()
		runtimeFatal("sync: WaitGroup.Add called from inside and outside synctest bubble")
	}

	if delta > 0 && wg.counter == 0 && wg.waiting != 0 {
		wg.lock.Unlock()
		panic("sync: WaitGroup misuse: Add called concurrently with Wait")
	}
	if delta > 0 && wg.counter > int(^uint32(0)>>1)-delta {
		wg.lock.Unlock()
		panic("sync: WaitGroup counter overflowed")
	}
	wg.counter += delta
	if wg.counter < 0 {
		wg.lock.Unlock()
		panic("sync: negative WaitGroup counter")
	}
	if wg.counter != 0 {
		wg.lock.Unlock()
		return
	}

	waiters := wg.waiters.Queue()
	if wg.waiting == 0 {
		wg.synctest = nil
	}
	wg.lock.Unlock()

	for waiter := waiters.Pop(); waiter != nil; waiter = waiters.Pop() {
		scheduleTask(waiter)
	}
}

func (wg *WaitGroup) addPlain(delta int) {
	if delta > 0 {
		for {
			counter := wg.futex.Load()
			if uint32(delta) > ^uint32(0)-counter {
				panic("sync: WaitGroup counter overflowed")
			}
			if wg.futex.CompareAndSwap(counter, counter+uint32(delta)) {
				return
			}
		}
	}

	for {
		counter := wg.futex.Load()
		if uint32(-delta) > counter {
			panic("sync: negative WaitGroup counter")
		}
		if !wg.futex.CompareAndSwap(counter, counter-uint32(-delta)) {
			continue
		}
		if counter-uint32(-delta) == 0 {
			wg.futex.WakeAll()
		}
		return
	}
}

func (wg *WaitGroup) Done() {
	wg.Add(-1)
}

func (wg *WaitGroup) Wait() {
	if !synctestIsEnabled() {
		wg.waitPlain()
		return
	}

	wg.lock.Lock()
	if wg.synctest == nil {
		wg.lock.Unlock()
		wg.waitPlain()
		return
	}
	current := task.Current()
	if wg.counter == 0 {
		if wg.waiting == 0 {
			wg.synctest = nil
		}
		wg.lock.Unlock()
		return
	}
	wg.waiting++
	wg.waiters.Push(current)
	if wg.synctest != nil && wg.synctest == current.SynctestBubble {
		synctestBlock(current)
	}
	wg.lock.Unlock()

	task.Pause()

	wg.lock.Lock()
	wg.waiting--
	if wg.counter != 0 {
		wg.lock.Unlock()
		panic("sync: WaitGroup is reused before previous Wait has returned")
	}
	if wg.waiting == 0 {
		wg.synctest = nil
	}
	wg.lock.Unlock()
}

func (wg *WaitGroup) waitPlain() {
	for {
		counter := wg.futex.Load()
		if counter == 0 {
			return
		}
		if wg.futex.Wait(counter) {
			return
		}
	}
}

func (wg *WaitGroup) Go(f func()) {
	wg.Add(1)
	go func() {
		defer wg.Done()
		f()
	}()
}

//go:linkname synctestBlock runtime.synctestBlock
func synctestBlock(*task.Task)

//go:linkname synctestIsEnabled runtime.synctestIsEnabled
func synctestIsEnabled() bool
