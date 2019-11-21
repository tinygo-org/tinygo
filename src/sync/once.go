package sync

type Once struct {
	done bool
	m    Mutex
}

func (o *Once) Do(f func()) {
	o.m.Lock()
	if o.done {
		o.m.Unlock()
		return
	}
	o.done = true
	f()
	o.m.Unlock()
}
