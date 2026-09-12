package runtime

import (
	"internal/task"
	"unsafe"
)

const synctestBaseTime = 946684800000000000

var synctestEnabled bool

func synctestIsEnabled() bool {
	return synctestEnabled
}

type synctestBubble struct {
	lock task.PMutex

	timers *timerNode

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

type synctestAssociation struct {
	next   *synctestAssociation
	ptr    unsafe.Pointer
	bubble *synctestBubble
}

var (
	synctestAssociationsLock task.PMutex
	synctestAssociations     *synctestAssociation
)

func taskSynctestBubble(t *task.Task) *synctestBubble {
	if t == nil || t.SynctestBubble == nil {
		return nil
	}
	return (*synctestBubble)(t.SynctestBubble)
}

func currentSynctestBubble() *synctestBubble {
	if !synctestEnabled {
		return nil
	}
	return taskSynctestBubble(task.Current())
}

func (bubble *synctestBubble) wakeLocked() *task.Task {
	if bubble.running != 0 || bubble.active != 0 {
		return nil
	}
	if bubble.timers != nil && bubble.timers.timer.when <= bubble.now {
		if bubble.rootSleeping {
			bubble.rootSleeping = false
			return bubble.root
		}
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
	return nil
}

func (bubble *synctestBubble) time() int64 {
	bubble.lock.Lock()
	now := bubble.now
	bubble.lock.Unlock()
	return now
}

func (bubble *synctestBubble) addTimer(tn *timerNode) {
	bubble.lock.Lock()
	if tn.timer.when <= bubble.now {
		bubble.lock.Unlock()
		tn.callback(tn, 0)
		return
	}
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
	bubble.lock.Unlock()
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
	return nil
}

func (bubble *synctestBubble) checkTimerAccess(op string) {
	if currentSynctestBubble() != bubble {
		runtimeFatal(op + " of synctest timer from outside bubble")
	}
}

func synctestWakeTaskTimer(tn *timerNode, delta int64) {
	scheduleTask(tn.timer.arg.(*task.Task))
}

func synctestSleep(duration int64) bool {
	if !synctestEnabled {
		return false
	}
	current := task.Current()
	bubble := taskSynctestBubble(current)
	if bubble == nil {
		return false
	}

	bubble.lock.Lock()
	when := bubble.now + duration
	bubble.lock.Unlock()
	tim := &timer{
		when:     when,
		arg:      current,
		synctest: bubble,
	}
	bubble.addTimer(&timerNode{
		timer:    tim,
		callback: synctestWakeTaskTimer,
	})
	synctestTaskBlock(current)
	task.Pause()
	return true
}

func synctestTaskCreated(t *task.Task) {
	if !synctestEnabled {
		return
	}
	bubble := taskSynctestBubble(t)
	bubble.lock.Lock()
	bubble.total++
	bubble.running++
	bubble.lock.Unlock()
}

func synctestTaskExited(t *task.Task) {
	if !synctestEnabled {
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
	if !synctestEnabled {
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
	if !synctestEnabled {
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
	if !synctestEnabled {
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

func synctestBlockBegin(t *task.Task) bool {
	return synctestTaskBlockBegin(t)
}

func synctestBlockEnd(t *task.Task, blocked bool) {
	synctestTaskBlockEnd(t, blocked)
}

func synctestBlock(t *task.Task) {
	synctestTaskBlock(t)
}

//go:linkname synctest_run internal/synctest.Run
func synctest_run(f func()) {
	synctestEnabled = true
	root := task.Current()
	if root.SynctestBubble != nil {
		panic("synctest.Run called from within a synctest bubble")
	}

	bubble := &synctestBubble{
		root: root,
		now:  synctestBaseTime,
	}

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
		if bubble.total == 0 {
			bubble.lock.Unlock()
			return
		}
		if bubble.running == 0 && bubble.active == 0 {
			if bubble.timers != nil && !bubble.done {
				timer := bubble.timers
				bubble.timers = timer.next
				timer.next = nil
				if timer.timer.when > bubble.now {
					bubble.now = timer.timer.when
				}
				bubble.lock.Unlock()

				root.SynctestBubble = unsafe.Pointer(bubble)
				timer.callback(timer, 0)
				root.SynctestBubble = nil
				continue
			}
			if bubble.waiter != nil {
				waiter := bubble.waiter
				bubble.waiter = nil
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
	bubble.waiting = false
	bubble.lock.Unlock()
}

//go:linkname synctest_isInBubble internal/synctest.IsInBubble
func synctest_isInBubble() bool {
	return currentSynctestBubble() != nil
}

//go:linkname synctest_associate internal/synctest.associate
func synctest_associate(p unsafe.Pointer) int {
	bubble := currentSynctestBubble()
	if bubble == nil {
		panic("goroutine is not in a bubble")
	}

	synctestAssociationsLock.Lock()
	for assoc := synctestAssociations; assoc != nil; assoc = assoc.next {
		if assoc.ptr == p {
			synctestAssociationsLock.Unlock()
			if assoc.bubble == bubble {
				return 1
			}
			return 2
		}
	}
	synctestAssociations = &synctestAssociation{
		next:   synctestAssociations,
		ptr:    p,
		bubble: bubble,
	}
	synctestAssociationsLock.Unlock()
	return 1
}

//go:linkname synctest_disassociate internal/synctest.disassociate
func synctest_disassociate(p unsafe.Pointer) {
	synctestAssociationsLock.Lock()
	for assoc := &synctestAssociations; *assoc != nil; assoc = &(*assoc).next {
		if (*assoc).ptr == p {
			*assoc = (*assoc).next
			break
		}
	}
	synctestAssociationsLock.Unlock()
}

//go:linkname synctest_isAssociated internal/synctest.isAssociated
func synctest_isAssociated(p unsafe.Pointer) bool {
	bubble := currentSynctestBubble()
	if bubble == nil {
		return false
	}

	synctestAssociationsLock.Lock()
	defer synctestAssociationsLock.Unlock()
	for assoc := synctestAssociations; assoc != nil; assoc = assoc.next {
		if assoc.ptr == p {
			return assoc.bubble == bubble
		}
	}
	return false
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
