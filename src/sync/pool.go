package sync

import "internal/task"

// Pool is a very simple implementation of sync.Pool.
type Pool struct {
	lock  task.PMutex
	New   func() interface{}
	items []interface{}
}

// Get returns an item in the pool, or the value of calling Pool.New() if there are no items.
func (p *Pool) Get() interface{} {
	p.lock.Lock()
	if len(p.items) > 0 {
		x := p.items[len(p.items)-1]
		p.items = p.items[:len(p.items)-1]
		p.lock.Unlock()
		return x
	}
	p.lock.Unlock()
	if p.New == nil {
		return nil
	}
	return p.New()
}

// Put adds a value back into the pool.
func (p *Pool) Put(x interface{}) {
	p.lock.Lock()
	p.items = append(p.items, x)
	p.lock.Unlock()
}
