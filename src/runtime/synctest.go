package runtime

import (
	"internal/task"
	"unsafe"
)

const synctestBaseTime = 946684800000000000

var synctestEnabled task.Uint32

func synctestIsEnabled() bool {
	return synctestEnabled.Load() != 0
}

type synctestBubble struct {
	lock synctestLock

	timers       *timerNode
	firingTimers *timerNode

	root   *task.Task
	main   *task.Task
	waiter *task.Task

	total   int
	running int
	active  int

	rootSleeping bool
	waiting      bool
	done         bool
	now          int64
	timerSeq     uint32
}

func taskSynctestBubble(t *task.Task) *synctestBubble {
	if t == nil || t.SynctestBubble == nil {
		return nil
	}
	return (*synctestBubble)(t.SynctestBubble)
}

func currentSynctestBubble() *synctestBubble {
	if !synctestIsEnabled() {
		return nil
	}
	return taskSynctestBubble(task.Current())
}

func (bubble *synctestBubble) wakeLocked() *task.Task {
	if bubble.running != 0 || bubble.active != 0 {
		return nil
	}
	bubble.active++
	if bubble.timers != nil && bubble.timers.timer.when <= bubble.now {
		if bubble.rootSleeping {
			bubble.rootSleeping = false
			return bubble.root
		}
		bubble.active--
		return nil
	}
	if bubble.waiter != nil {
		waiter := bubble.waiter
		bubble.waiter = nil
		return waiter
	}
	if bubble.rootSleeping {
		bubble.rootSleeping = false
		return bubble.root
	}
	bubble.active--
	return nil
}

func (bubble *synctestBubble) addFiringTimerLocked(tn *timerNode) {
	tn.stopped = false
	tn.firingNext = bubble.firingTimers
	bubble.firingTimers = tn
}

func (bubble *synctestBubble) removeFiringTimerLocked(tn *timerNode) {
	for queue := &bubble.firingTimers; *queue != nil; queue = &(*queue).firingNext {
		if *queue == tn {
			*queue = tn.firingNext
			tn.firingNext = nil
			return
		}
	}
}

func (bubble *synctestBubble) stopFiringTimerLocked(tim *timer) {
	for tn := bubble.firingTimers; tn != nil; tn = tn.firingNext {
		if tn.timer == tim {
			tn.stopped = true
			return
		}
	}
}

func (bubble *synctestBubble) time() int64 {
	bubble.lock.Lock()
	now := bubble.now
	bubble.lock.Unlock()
	return now
}

func (bubble *synctestBubble) addTimer(tn *timerNode) {
	if bubble.queueTimer(tn) {
		tn.callback(tn, 0)
	}
}

func (bubble *synctestBubble) queueTimer(tn *timerNode) bool {
	bubble.lock.Lock()
	if tn.timer.when <= bubble.now {
		bubble.addFiringTimerLocked(tn)
		bubble.lock.Unlock()
		return true
	}
	bubble.addTimerLocked(tn)
	bubble.lock.Unlock()
	return false
}

func (bubble *synctestBubble) finishTimer(tn *timerNode) {
	bubble.lock.Lock()
	bubble.removeFiringTimerLocked(tn)
	if tn.stopped {
		bubble.lock.Unlock()
		return
	}
	if tn.timer.period == 0 {
		bubble.lock.Unlock()
		return
	}
	next := tn.timer.when + tn.timer.period
	if next < 0 {
		next = 1<<63 - 1
	}
	tn.timer.when = next
	bubble.addTimerLocked(tn)
	bubble.lock.Unlock()
}

func (bubble *synctestBubble) addTimerLocked(tn *timerNode) {
	bubble.timerSeq++
	insertBeforeEqual := (bubble.timerSeq/2)&1 != 0
	queue := &bubble.timers
	for *queue != nil {
		if (*queue).timer.when > tn.timer.when {
			break
		}
		if insertBeforeEqual && (*queue).timer.when == tn.timer.when {
			break
		}
		queue = &(*queue).next
	}
	tn.next = *queue
	*queue = tn
}

func (bubble *synctestBubble) removeTimer(tim *timer) *timerNode {
	bubble.lock.Lock()
	defer bubble.lock.Unlock()
	for queue := &bubble.timers; *queue != nil; queue = &(*queue).next {
		if (*queue).timer == tim {
			node := *queue
			*queue = node.next
			node.next = nil
			return node
		}
	}
	bubble.stopFiringTimerLocked(tim)
	return nil
}

func (bubble *synctestBubble) checkTimerAccess(op string) {
	if currentSynctestBubble() != bubble {
		runtimeFatal(op + " of synctest timer from outside bubble")
	}
}

func synctestWakeTaskTimer(tn *timerNode, delta int64) {
	tn.timer.synctest.finishTimer(tn)
	scheduleTask(tn.timer.arg.(*task.Task))
}

func synctestSleep(duration int64) bool {
	if !synctestIsEnabled() {
		return false
	}
	current := task.Current()
	bubble := taskSynctestBubble(current)
	if bubble == nil {
		return false
	}

	bubble.lock.Lock()
	bubble.active++
	when := bubble.now + duration
	if when < 0 {
		when = 1<<63 - 1
	}
	bubble.lock.Unlock()
	tim := &timer{
		when:     when,
		arg:      current,
		synctest: bubble,
	}
	node := &timerNode{
		timer:    tim,
		callback: synctestWakeTaskTimer,
	}
	runNow := bubble.queueTimer(node)
	bubble.lock.Lock()
	if runNow {
		bubble.removeFiringTimerLocked(node)
		bubble.active--
		if bubble.active < 0 {
			bubble.lock.Unlock()
			runtimeFatal("synctest: invalid sleep transition")
		}
		bubble.lock.Unlock()
		return true
	}
	if !current.SynctestBlocked {
		current.SynctestBlocked = true
		bubble.running--
	}
	bubble.active--
	if bubble.running < 0 || bubble.active < 0 {
		bubble.lock.Unlock()
		runtimeFatal("synctest: invalid sleep transition")
	}
	wake := bubble.wakeLocked()
	bubble.lock.Unlock()
	if wake != nil {
		scheduleTask(wake)
	}
	task.Pause()
	return true
}

func synctestTaskCreated(t *task.Task) {
	if !synctestIsEnabled() {
		return
	}
	bubble := taskSynctestBubble(t)
	bubble.lock.Lock()
	bubble.total++
	bubble.running++
	bubble.lock.Unlock()
}

func synctestTaskExited(t *task.Task) {
	if !synctestIsEnabled() {
		return
	}
	bubble := taskSynctestBubble(t)
	bubble.lock.Lock()
	if t.SynctestBlocked {
		t.SynctestBlocked = false
	} else {
		bubble.running--
	}
	bubble.total--
	if t == bubble.main {
		bubble.done = true
	}
	if bubble.running < 0 || bubble.total < 0 {
		bubble.lock.Unlock()
		runtimeFatal("synctest: invalid task count")
	}
	wake := bubble.wakeLocked()
	bubble.lock.Unlock()
	if wake != nil {
		scheduleTask(wake)
	}
}

func synctestTaskWake(t *task.Task) {
	if !synctestIsEnabled() {
		return
	}
	bubble := taskSynctestBubble(t)
	if bubble == nil {
		return
	}
	bubble.lock.Lock()
	if t.SynctestBlocked {
		t.SynctestBlocked = false
		bubble.running++
	}
	bubble.lock.Unlock()
}

func synctestTaskBlock(t *task.Task) {
	if !synctestIsEnabled() {
		return
	}
	bubble := taskSynctestBubble(t)
	if bubble == nil {
		return
	}
	bubble.lock.Lock()
	if !t.SynctestBlocked {
		t.SynctestBlocked = true
		bubble.running--
	}
	if bubble.running < 0 {
		bubble.lock.Unlock()
		runtimeFatal("synctest: invalid running task count")
	}
	wake := bubble.wakeLocked()
	bubble.lock.Unlock()
	if wake != nil {
		scheduleTask(wake)
	}
}

func synctestTaskBlockBegin(t *task.Task) bool {
	if !synctestIsEnabled() {
		return false
	}
	bubble := taskSynctestBubble(t)
	if bubble == nil {
		return false
	}
	bubble.lock.Lock()
	bubble.active++
	bubble.lock.Unlock()
	return true
}

func synctestTaskBlockEnd(t *task.Task, blocked bool) {
	bubble := taskSynctestBubble(t)
	if bubble == nil {
		return
	}
	bubble.lock.Lock()
	if blocked && !t.SynctestBlocked {
		t.SynctestBlocked = true
		bubble.running--
	}
	bubble.active--
	if bubble.running < 0 || bubble.active < 0 {
		bubble.lock.Unlock()
		runtimeFatal("synctest: invalid block transition")
	}
	wake := bubble.wakeLocked()
	bubble.lock.Unlock()
	if wake != nil {
		scheduleTask(wake)
	}
}

func synctestTaskBlockCommit(t *task.Task, state *task.Uint32, old, new uint32) bool {
	bubble := taskSynctestBubble(t)
	if bubble == nil {
		return false
	}
	bubble.lock.Lock()
	blocked := state.CompareAndSwap(old, new)
	if blocked && !t.SynctestBlocked {
		t.SynctestBlocked = true
		bubble.running--
	}
	bubble.active--
	if bubble.running < 0 || bubble.active < 0 {
		bubble.lock.Unlock()
		runtimeFatal("synctest: invalid block transition")
	}
	wake := bubble.wakeLocked()
	bubble.lock.Unlock()
	if wake != nil {
		scheduleTask(wake)
	}
	return blocked
}

func synctestBlockBegin(t *task.Task) bool {
	return synctestTaskBlockBegin(t)
}

func synctestBlockEnd(t *task.Task, blocked bool) {
	synctestTaskBlockEnd(t, blocked)
}

func synctestBlockCommit(t *task.Task, state *task.Uint32, old, new uint32) bool {
	return synctestTaskBlockCommit(t, state, old, new)
}

func synctestBlock(t *task.Task) {
	synctestTaskBlock(t)
}

//go:linkname synctest_run internal/synctest.Run
func synctest_run(f func()) {
	synctestEnabled.Store(1)
	root := task.Current()
	if root.SynctestBubble != nil {
		panic("synctest.Run called from within a synctest bubble")
	}

	bubble := &synctestBubble{
		root: root,
		now:  synctestBaseTime,
	}

	// Let the new goroutine inherit the bubble from the root task.
	root.SynctestBubble = unsafe.Pointer(bubble)
	go func() {
		bubble.lock.Lock()
		bubble.main = task.Current()
		bubble.lock.Unlock()
		f()
	}()
	root.SynctestBubble = nil

	for {
		bubble.lock.Lock()
		if bubble.rootSleeping {
			bubble.lock.Unlock()
			runtimeFatal("synctest: root resumed while marked sleeping")
		}
		if bubble.total == 0 && bubble.active == 0 {
			bubble.lock.Unlock()
			return
		}
		if bubble.running == 0 && bubble.active == 0 {
			dueTimer := bubble.timers != nil && bubble.timers.timer.when <= bubble.now
			if bubble.timers != nil && !bubble.done && (dueTimer || bubble.waiter == nil) {
				timer := bubble.timers
				bubble.timers = timer.next
				timer.next = nil
				bubble.addFiringTimerLocked(timer)
				bubble.active++
				if timer.timer.when > bubble.now {
					bubble.now = timer.timer.when
				}
				bubble.lock.Unlock()

				// Timer callbacks run on the root task inside the bubble.
				root.SynctestBubble = unsafe.Pointer(bubble)
				timer.callback(timer, 0)
				root.SynctestBubble = nil
				bubble.lock.Lock()
				bubble.active--
				if bubble.active < 0 {
					bubble.lock.Unlock()
					runtimeFatal("synctest: invalid active count")
				}
				bubble.lock.Unlock()
				continue
			}
			if bubble.waiter != nil {
				waiter := bubble.waiter
				bubble.waiter = nil
				bubble.active++
				bubble.lock.Unlock()
				scheduleTask(waiter)
				continue
			}
			done := bubble.done
			bubble.lock.Unlock()
			if done {
				panic("deadlock: main bubble goroutine has exited but blocked goroutines remain")
			}
			panic("deadlock: all goroutines in bubble are blocked")
		}
		bubble.rootSleeping = true
		bubble.lock.Unlock()
		task.Pause()
		bubble.lock.Lock()
		bubble.active--
		if bubble.active < 0 {
			bubble.lock.Unlock()
			runtimeFatal("synctest: invalid active count")
		}
		bubble.lock.Unlock()
	}
}

//go:linkname synctest_wait internal/synctest.Wait
func synctest_wait() {
	current := task.Current()
	bubble := taskSynctestBubble(current)
	if bubble == nil {
		panic("goroutine is not in a bubble")
	}

	bubble.lock.Lock()
	if bubble.waiting {
		bubble.lock.Unlock()
		panic("wait already in progress")
	}
	bubble.waiting = true
	current.SynctestBlocked = true
	bubble.running--
	dueTimer := bubble.timers != nil && bubble.timers.timer.when <= bubble.now
	if bubble.running == 0 && bubble.active == 0 && !dueTimer {
		current.SynctestBlocked = false
		bubble.running++
		bubble.waiting = false
		bubble.lock.Unlock()
		return
	}
	bubble.waiter = current
	wake := bubble.wakeLocked()
	bubble.lock.Unlock()
	if wake != nil {
		scheduleTask(wake)
	}

	task.Pause()

	bubble.lock.Lock()
	bubble.active--
	if bubble.active < 0 {
		bubble.lock.Unlock()
		runtimeFatal("synctest: invalid active count")
	}
	bubble.waiting = false
	bubble.lock.Unlock()
}

//go:linkname synctest_isInBubble internal/synctest.IsInBubble
func synctest_isInBubble() bool {
	return currentSynctestBubble() != nil
}

//go:linkname synctest_acquire internal/synctest.acquire
func synctest_acquire() any {
	bubble := currentSynctestBubble()
	if bubble == nil {
		return nil
	}
	bubble.lock.Lock()
	bubble.active++
	bubble.lock.Unlock()
	return bubble
}

//go:linkname synctest_release internal/synctest.release
func synctest_release(value any) {
	bubble := value.(*synctestBubble)
	bubble.lock.Lock()
	bubble.active--
	if bubble.active < 0 {
		bubble.lock.Unlock()
		runtimeFatal("synctest: invalid active count")
	}
	wake := bubble.wakeLocked()
	bubble.lock.Unlock()
	if wake != nil {
		scheduleTask(wake)
	}
}

//go:linkname synctest_inBubble internal/synctest.inBubble
func synctest_inBubble(value any, f func()) {
	current := task.Current()
	if current.SynctestBubble != nil {
		panic("goroutine is already bubbled")
	}
	current.SynctestBubble = unsafe.Pointer(value.(*synctestBubble))
	defer func() {
		current.SynctestBubble = nil
	}()
	f()
}
