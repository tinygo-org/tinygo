package sync

// Pool is a very simple implementation of sync.Pool.
type Pool struct {
	New   func() interface{}
	items []interface{}
}

// Get returns the value of calling Pool.New().
func (p *Pool) Get() interface{} {
	if len(p.items) > 0 {
		x := p.items[len(p.items)-1]
		p.items = p.items[:len(p.items)-1]
		return x
	}
	if p.New == nil {
		return nil
	}
	return p.New()
}

// Put drops the value put into the pool.
func (p *Pool) Put(x interface{}) {
	p.items = append(p.items, x)
}
