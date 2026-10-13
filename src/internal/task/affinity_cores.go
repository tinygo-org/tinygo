//go:build scheduler.cores

package task

// Affinity holds the assigned core of a task.
type Affinity struct {
	// Core is the assigned core plus one, or 0 when the task may run on any core.
	Core uint8
}

// CanRunOn reports whether the task may run on the given core.
func (t *Task) CanRunOn(core uint32) bool {
	return t.Affinity.Core == 0 || uint32(t.Affinity.Core-1) == core
}

// PopCore removes and returns the first task that may run on the given core.
func (q *Queue) PopCore(core uint32) *Task {
	mask := lockAtomics()
	var prev *Task
	for t := q.head; t != nil; prev, t = t, t.Next {
		if !t.CanRunOn(core) {
			continue
		}
		if prev == nil {
			q.head = t.Next
		} else {
			prev.Next = t.Next
		}
		if q.tail == t {
			q.tail = prev
		}
		t.Next = nil
		unlockAtomics(mask)
		return t
	}
	unlockAtomics(mask)
	return nil
}
