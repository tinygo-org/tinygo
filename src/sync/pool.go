package sync

import (
	"internal/task"
	_ "unsafe"
)

// Pool is a very simple implementation of sync.Pool. Like the standard
// runtime it drops items after two collections, so a burst is not kept forever.
type Pool struct {
	lock   task.PMutex
	New    func() interface{}
	items  []interface{}
	victim []interface{}
	gcSeen uint32
}

//go:linkname runtime_poolGCCount runtime.poolGCCount
func runtime_poolGCCount() uint32

// age moves the items to the victim list after a collection and drops both
// after two. The caller holds p.lock.
func (p *Pool) age() {
	n := runtime_poolGCCount()
	if n == p.gcSeen {
		return
	}
	if n-p.gcSeen == 1 {
		p.victim = p.items
	} else {
		p.victim = nil
	}
	p.items = nil
	p.gcSeen = n
}

// Get returns an item in the pool, or the value of calling Pool.New() if there are no items.
func (p *Pool) Get() interface{} {
	p.lock.Lock()
	p.age()
	x, ok := pop(&p.items)
	if !ok {
		x, ok = pop(&p.victim)
	}
	p.lock.Unlock()
	if ok {
		return x
	}
	if p.New == nil {
		return nil
	}
	return p.New()
}

// Put adds a value back into the pool.
func (p *Pool) Put(x interface{}) {
	if x == nil {
		return
	}
	p.lock.Lock()
	p.age()
	p.items = append(p.items, x)
	p.lock.Unlock()
}

func pop(l *[]interface{}) (interface{}, bool) {
	s := *l
	if len(s) == 0 {
		return nil, false
	}
	x := s[len(s)-1]
	s[len(s)-1] = nil
	*l = s[:len(s)-1]
	return x, true
}
