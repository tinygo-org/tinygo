package sync

// Pool is a very simple implementation of sync.Pool. It does not actually
// implement a pool.
type Pool struct {
	New func() interface{}
}

// Get returns the value of calling Pool.New().
func (p *Pool) Get() interface{} {
	if p.New == nil {
		return nil
	}
	return p.New()
}

// Put drops the value put into the pool.
func (p *Pool) Put(x interface{}) {
}
